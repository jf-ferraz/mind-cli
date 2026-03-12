package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/tui/components"
)

// IterationsView is Tab 3: iteration table with type filter and detail expander.
type IterationsView struct {
	viewState     ViewState
	iterations    []domain.Iteration
	filtered      []domain.Iteration
	activeType    *domain.RequestType
	cursor        int
	expandedIndex int
	width         int
	height        int
}

// NewIterationsView creates an IterationsView.
func NewIterationsView() IterationsView {
	return IterationsView{
		viewState:     ViewLoading,
		expandedIndex: -1,
	}
}

func (v IterationsView) Init() tea.Cmd { return nil }

func (v IterationsView) Update(msg tea.Msg) (IterationsView, tea.Cmd) {
	switch msg := msg.(type) {
	case iterationsLoadedMsg:
		v.iterations = msg.iterations
		if len(v.iterations) == 0 {
			v.viewState = ViewEmpty
		} else {
			v.viewState = ViewReady
		}
		v.applyFilter()

	case iterationsErrorMsg:
		v.viewState = ViewError

	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		return v.handleKey(msg)
	}
	return v, nil
}

func (v IterationsView) handleKey(msg tea.KeyMsg) (IterationsView, tea.Cmd) {
	switch msg.String() {
	case "a":
		v.activeType = nil
		v.applyFilter()
	case "n":
		t := domain.TypeNewProject
		v.activeType = &t
		v.applyFilter()
	case "e":
		t := domain.TypeEnhancement
		v.activeType = &t
		v.applyFilter()
	case "b":
		t := domain.TypeBugFix
		v.activeType = &t
		v.applyFilter()
	case "r":
		// 'r' is type filter for REFACTOR in iterations tab context
		t := domain.TypeRefactor
		v.activeType = &t
		v.applyFilter()
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.filtered)-1 {
			v.cursor++
		}
	case "enter":
		if v.expandedIndex == v.cursor {
			v.expandedIndex = -1
		} else {
			v.expandedIndex = v.cursor
		}
	}
	return v, nil
}

func (v *IterationsView) applyFilter() {
	v.filtered = nil
	for _, iter := range v.iterations {
		if v.activeType != nil && iter.Type != *v.activeType {
			continue
		}
		v.filtered = append(v.filtered, iter)
	}
	if v.cursor >= len(v.filtered) {
		v.cursor = len(v.filtered) - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
	v.expandedIndex = -1
}

func (v IterationsView) View() string {
	switch v.viewState {
	case ViewLoading:
		return components.EmptyState("Loading iterations...", v.width, v.height)
	case ViewError:
		return components.EmptyState("Error loading iterations.\nPress r to retry", v.width, v.height)
	case ViewEmpty:
		return components.EmptyState("No iterations yet. Start a workflow to create one.", v.width, v.height)
	}

	// Type filter bar
	var filterParts []string
	types := []struct {
		key  string
		typ  *domain.RequestType
		name string
	}{
		{"a", nil, "all"},
		{"n", reqPtr(domain.TypeNewProject), "NEW"},
		{"e", reqPtr(domain.TypeEnhancement), "ENHANCE"},
		{"b", reqPtr(domain.TypeBugFix), "BUGFIX"},
		{"r", reqPtr(domain.TypeRefactor), "REFACTOR"},
	}
	for _, t := range types {
		active := v.activeType == nil && t.typ == nil
		if t.typ != nil && v.activeType != nil && *t.typ == *v.activeType {
			active = true
		}
		if active {
			filterParts = append(filterParts, theme.TabActive.Render("["+t.key+" "+t.name+"]"))
		} else {
			filterParts = append(filterParts, theme.Dim.Render(" "+t.key+" "+t.name+" "))
		}
	}

	var b strings.Builder
	b.WriteString(strings.Join(filterParts, " ") + "\n\n")

	// Table header
	header := fmt.Sprintf("  %-5s %-14s %-28s %-12s %s", "#", "Type", "Name", "Status", "Files")
	b.WriteString(theme.Heading.Render(header) + "\n")
	b.WriteString("  " + strings.Repeat("─", v.width-4) + "\n")

	for i, iter := range v.filtered {
		present := 0
		for _, a := range iter.Artifacts {
			if a.Exists {
				present++
			}
		}

		statusIcon := theme.Dim.Render("○")
		switch iter.Status {
		case domain.IterComplete:
			statusIcon = theme.Pass.Render("✓")
		case domain.IterInProgress:
			statusIcon = theme.Warn.Render("▸")
		}

		typeStyle := theme.Dim
		switch iter.Type {
		case domain.TypeNewProject:
			typeStyle = theme.TypeNew
		case domain.TypeEnhancement:
			typeStyle = theme.TypeEnhancement
		case domain.TypeBugFix:
			typeStyle = theme.TypeBugFix
		case domain.TypeRefactor:
			typeStyle = theme.TypeRefactor
		}

		line := fmt.Sprintf("  %03d   %s %-28s %s %-10s %d/%d",
			iter.Seq,
			typeStyle.Render(fmt.Sprintf("%-14s", string(iter.Type))),
			truncStr(iter.Descriptor, 28),
			statusIcon,
			string(iter.Status),
			present,
			len(domain.ExpectedArtifacts))

		if i == v.cursor {
			b.WriteString(theme.Selected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}

		// Expanded detail
		if i == v.expandedIndex {
			for _, a := range iter.Artifacts {
				icon := theme.Pass.Render("✓")
				if !a.Exists {
					icon = theme.Fail.Render("✗")
				}
				b.WriteString(fmt.Sprintf("        %s %s\n", icon, a.Name))
			}
		}
	}

	return b.String()
}

func reqPtr(t domain.RequestType) *domain.RequestType { return &t }

func truncStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
