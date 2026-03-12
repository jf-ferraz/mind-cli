package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/deps"
	"github.com/jf-ferraz/mind-cli/tui/components"
)

// DocsView is Tab 2: document list with zone filter, search, and preview.
type DocsView struct {
	deps           *deps.Deps
	viewState      ViewState
	documents      []domain.Document
	filtered       []domain.Document
	activeZone     *domain.Zone
	searchQuery    string
	searchActive   bool
	cursor         int
	previewVisible bool
	previewContent string
	width          int
	height         int
}

// NewDocsView creates a DocsView.
func NewDocsView(deps *deps.Deps) DocsView {
	return DocsView{
		deps:      deps,
		viewState: ViewLoading,
	}
}

// Init returns nil (data propagated from health load).
func (v DocsView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the Docs tab.
func (v DocsView) Update(msg tea.Msg) (DocsView, tea.Cmd) {
	switch msg := msg.(type) {
	case healthLoadedMsg:
		if msg.health == nil {
			v.viewState = ViewEmpty
			return v, nil
		}
		v.documents = nil
		for _, zone := range domain.AllZones {
			if zh, ok := msg.health.Zones[zone]; ok {
				v.documents = append(v.documents, zh.Files...)
			}
		}
		if len(v.documents) == 0 {
			v.viewState = ViewEmpty
		} else {
			v.viewState = ViewReady
		}
		v.applyFilters()

	case healthErrorMsg:
		v.viewState = ViewError

	case previewLoadedMsg:
		v.previewContent = msg.content
		v.previewVisible = true

	case previewErrorMsg:
		v.previewContent = "Error loading preview: " + msg.err.Error()
		v.previewVisible = true

	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		if v.searchActive {
			return v.handleSearchKey(msg)
		}
		return v.handleNormalKey(msg)
	}
	return v, nil
}

func (v DocsView) handleSearchKey(msg tea.KeyMsg) (DocsView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.searchActive = false
		v.searchQuery = ""
		v.applyFilters()
	case "backspace":
		if len(v.searchQuery) > 0 {
			v.searchQuery = v.searchQuery[:len(v.searchQuery)-1]
			v.applyFilters()
		}
	case "enter":
		v.searchActive = false
	default:
		if len(msg.String()) == 1 {
			v.searchQuery += msg.String()
			v.applyFilters()
			v.cursor = 0
		}
	}
	return v, nil
}

func (v DocsView) handleNormalKey(msg tea.KeyMsg) (DocsView, tea.Cmd) {
	switch msg.String() {
	case "/":
		v.searchActive = true
	case "esc":
		if v.previewVisible {
			v.previewVisible = false
		}
	case "a":
		v.activeZone = nil
		v.applyFilters()
	case "s":
		z := domain.ZoneSpec
		v.activeZone = &z
		v.applyFilters()
	case "b":
		z := domain.ZoneBlueprints
		v.activeZone = &z
		v.applyFilters()
	case "t":
		z := domain.ZoneState
		v.activeZone = &z
		v.applyFilters()
	case "i":
		z := domain.ZoneIterations
		v.activeZone = &z
		v.applyFilters()
	case "k":
		z := domain.ZoneKnowledge
		v.activeZone = &z
		v.applyFilters()
	case "up":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.filtered)-1 {
			v.cursor++
		}
	case "enter":
		if len(v.filtered) > 0 && v.cursor < len(v.filtered) {
			return v, v.loadPreview(v.filtered[v.cursor].Path)
		}
	case "e":
		if len(v.filtered) > 0 && v.cursor < len(v.filtered) {
			return v, v.openEditor(v.filtered[v.cursor].Path)
		}
	}
	return v, nil
}

func (v *DocsView) applyFilters() {
	v.filtered = nil
	query := strings.ToLower(v.searchQuery)
	for _, doc := range v.documents {
		if v.activeZone != nil && doc.Zone != *v.activeZone {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(doc.Name), query) {
			continue
		}
		v.filtered = append(v.filtered, doc)
	}
	if v.cursor >= len(v.filtered) {
		v.cursor = len(v.filtered) - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
}

func (v DocsView) loadPreview(relPath string) tea.Cmd {
	return func() tea.Msg {
		content, err := v.deps.DocRepo.Read(relPath)
		if err != nil {
			return previewErrorMsg{err: err}
		}
		return previewLoadedMsg{content: string(content)}
	}
}

func (v DocsView) openEditor(relPath string) tea.Cmd {
	cmd, err := editorCmd(relPath)
	if err != nil {
		return func() tea.Msg {
			return previewErrorMsg{err: err}
		}
	}
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return nil
	})
}

// View renders the Docs tab.
func (v DocsView) View() string {
	switch v.viewState {
	case ViewLoading:
		return components.EmptyState("Loading documents...", v.width, v.height)
	case ViewError:
		return components.EmptyState("Error loading documents.\nPress r to retry", v.width, v.height)
	case ViewEmpty:
		return components.EmptyState("No documents found.", v.width, v.height)
	}

	contentWidth := v.width - 2

	// Zone filter bar
	var filterParts []string
	zones := []struct {
		key  string
		zone *domain.Zone
		name string
	}{
		{"a", nil, "all"},
		{"s", zonePtr(domain.ZoneSpec), "spec"},
		{"b", zonePtr(domain.ZoneBlueprints), "blueprints"},
		{"t", zonePtr(domain.ZoneState), "state"},
		{"i", zonePtr(domain.ZoneIterations), "iterations"},
		{"k", zonePtr(domain.ZoneKnowledge), "knowledge"},
	}
	for _, z := range zones {
		active := v.activeZone == nil && z.zone == nil
		if z.zone != nil && v.activeZone != nil && *z.zone == *v.activeZone {
			active = true
		}
		if active {
			filterParts = append(filterParts, theme.TabActive.Render("["+z.key+" "+z.name+"]"))
		} else {
			filterParts = append(filterParts, theme.Dim.Render(" "+z.key+" "+z.name+" "))
		}
	}
	filterBar := strings.Join(filterParts, " ")

	// Search indicator
	searchBar := ""
	if v.searchActive {
		searchBar = "\n  Search: " + v.searchQuery + "█"
	} else if v.searchQuery != "" {
		searchBar = "\n  Filter: " + v.searchQuery
	}

	var b strings.Builder
	b.WriteString(filterBar + searchBar + "\n\n")

	if v.previewVisible {
		return b.String() + v.renderWithPreview(contentWidth)
	}

	// Document list
	currentZone := domain.Zone("")
	for i, doc := range v.filtered {
		if doc.Zone != currentZone {
			currentZone = doc.Zone
			b.WriteString(theme.Heading.Render("  "+string(doc.Zone)+"/") + "\n")
		}

		indicator := theme.Pass.Render("✓")
		if doc.IsStub {
			indicator = theme.Fail.Render("✗")
		}
		line := fmt.Sprintf("  %s %-30s %6d bytes  %s",
			indicator,
			doc.Name+".md",
			doc.Size,
			doc.ModTime.Format("2006-01-02"))

		if i == v.cursor {
			b.WriteString(theme.Selected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

func (v DocsView) renderWithPreview(contentWidth int) string {
	listWidth := contentWidth * 40 / 100
	previewWidth := contentWidth - listWidth

	var listBuf strings.Builder
	currentZone := domain.Zone("")
	for i, doc := range v.filtered {
		if doc.Zone != currentZone {
			currentZone = doc.Zone
			listBuf.WriteString("  " + string(doc.Zone) + "/\n")
		}
		name := doc.Name + ".md"
		if len(name) > listWidth-6 {
			name = name[:listWidth-9] + "..."
		}
		line := fmt.Sprintf("  %s", name)
		if i == v.cursor {
			listBuf.WriteString(theme.Selected.Render(line) + "\n")
		} else {
			listBuf.WriteString(line + "\n")
		}
	}

	listCol := lipgloss.NewStyle().Width(listWidth).Render(listBuf.String())
	previewCol := lipgloss.NewStyle().Width(previewWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Render(truncateLines(v.previewContent, v.height-8))

	return lipgloss.JoinHorizontal(lipgloss.Top, listCol, previewCol)
}

func zonePtr(z domain.Zone) *domain.Zone { return &z }

func truncateLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines && maxLines > 0 {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}
