package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/jf-ferraz/mind-cli/internal/deps"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
)

// App is the top-level Bubble Tea model for the TUI dashboard.
type App struct {
	deps      *deps.Deps
	width     int
	height    int
	activeTab TabID
	status    StatusView
	docs      DocsView
	iters     IterationsView
	checks    ChecksView
	quality   QualityView
	help      HelpModel
	showHelp  bool
	loaded    bool

	// Project info for title bar
	projectName string
	gitBranch   string
	version     string
}

// NewApp constructs the TUI application with injected dependencies.
func NewApp(deps *deps.Deps, version string) App {
	// Detect project info
	name := ""
	if project, err := fs.DetectProject(deps.ProjectRoot); err == nil {
		name = project.Name
	}

	branch := detectBranch(deps.ProjectRoot)

	return App{
		deps:        deps,
		activeTab:   TabStatus,
		status:      NewStatusView(),
		docs:        NewDocsView(deps),
		iters:       NewIterationsView(),
		checks:      NewChecksView(deps),
		quality:     NewQualityView(),
		help:        NewHelpModel(),
		projectName: name,
		gitBranch:   branch,
		version:     version,
	}
}

// Init dispatches initial data loading commands.
func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.loadHealth(),
		a.loadIterations(),
		a.loadQuality(),
		// Validation NOT loaded on init (FR-122: lazy)
	)
}

// Update processes messages and delegates to the active tab.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.SetSize(msg.Width, msg.Height)

		// Propagate to all tabs
		a.status, _ = a.status.Update(msg)
		a.docs, _ = a.docs.Update(msg)
		a.iters, _ = a.iters.Update(msg)
		a.checks, _ = a.checks.Update(msg)
		a.quality, _ = a.quality.Update(msg)
		return a, nil

	case tea.KeyMsg:
		// Ctrl+C always force-quits
		if key.Matches(msg, globalKeys.ForceQuit) {
			return a, tea.Quit
		}

		// Help overlay intercepts keys when open
		if a.showHelp {
			switch {
			case key.Matches(msg, globalKeys.Help),
				msg.String() == "esc",
				msg.String() == "q":
				a.showHelp = false
				return a, nil
			}
			return a, nil
		}

		// Global keys
		switch {
		case key.Matches(msg, globalKeys.Help):
			a.showHelp = true
			return a, nil
		case key.Matches(msg, globalKeys.Quit):
			return a, tea.Quit
		case key.Matches(msg, globalKeys.Tab1):
			return a.switchTab(TabStatus)
		case key.Matches(msg, globalKeys.Tab2):
			return a.switchTab(TabDocs)
		case key.Matches(msg, globalKeys.Tab3):
			return a.switchTab(TabIterations)
		case key.Matches(msg, globalKeys.Tab4):
			return a.switchTab(TabChecks)
		case key.Matches(msg, globalKeys.Tab5):
			return a.switchTab(TabQuality)
		case key.Matches(msg, globalKeys.NextTab):
			next := TabID((int(a.activeTab) + 1) % TabCount)
			return a.switchTab(next)
		case key.Matches(msg, globalKeys.PrevTab):
			prev := TabID((int(a.activeTab) - 1 + TabCount) % TabCount)
			return a.switchTab(prev)
		case key.Matches(msg, globalKeys.Refresh):
			// Let 'r' pass through to tab-specific handlers:
			// - Docs: search input active (existing)
			// - Iterations: 'r' toggles REFACTOR type filter (FR-108)
			// - Checks: 'r' re-runs validation (FR-111/FR-112)
			if a.activeTab == TabDocs && a.docs.searchActive {
				break // fall through to tab handler
			}
			if a.activeTab == TabIterations || a.activeTab == TabChecks {
				break // fall through to tab handler
			}
			return a, tea.Batch(a.loadHealth(), a.loadIterations(), a.loadQuality())
		}

		// Delegate to active tab
		return a.delegateKey(msg)

	// Route data messages to appropriate tabs
	case healthLoadedMsg:
		a.loaded = true
		var cmd1, cmd2 tea.Cmd
		a.status, cmd1 = a.status.Update(msg)
		a.docs, cmd2 = a.docs.Update(msg)
		return a, tea.Batch(cmd1, cmd2)

	case healthErrorMsg:
		a.status, _ = a.status.Update(msg)
		a.docs, _ = a.docs.Update(msg)
		return a, nil

	case iterationsLoadedMsg:
		a.iters, _ = a.iters.Update(msg)
		return a, nil

	case iterationsErrorMsg:
		a.iters, _ = a.iters.Update(msg)
		return a, nil

	case qualityLoadedMsg:
		a.quality, _ = a.quality.Update(msg)
		return a, nil

	case qualityErrorMsg:
		a.quality, _ = a.quality.Update(msg)
		return a, nil

	case validationCompleteMsg, validationErrorMsg:
		var cmd tea.Cmd
		a.checks, cmd = a.checks.Update(msg)
		return a, cmd

	case previewLoadedMsg, previewErrorMsg:
		var cmd tea.Cmd
		a.docs, cmd = a.docs.Update(msg)
		return a, cmd
	}

	// Spinner ticks for checks tab
	if a.activeTab == TabChecks {
		var cmd tea.Cmd
		a.checks, cmd = a.checks.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a App) switchTab(tab TabID) (tea.Model, tea.Cmd) {
	a.activeTab = tab
	// Notify the tab it was activated (for lazy loading)
	var cmd tea.Cmd
	if tab == TabChecks {
		a.checks, cmd = a.checks.Update(tabActivatedMsg{tab: tab})
	}
	return a, cmd
}

func (a App) delegateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.activeTab {
	case TabStatus:
		a.status, cmd = a.status.Update(msg)
	case TabDocs:
		a.docs, cmd = a.docs.Update(msg)
	case TabIterations:
		a.iters, cmd = a.iters.Update(msg)
	case TabChecks:
		a.checks, cmd = a.checks.Update(msg)
	case TabQuality:
		a.quality, cmd = a.quality.Update(msg)
	}
	return a, cmd
}

// View renders the full TUI screen.
func (a App) View() string {
	// Terminal too small check
	if a.width < MinWidth || a.height < MinHeight {
		msg := fmt.Sprintf("Terminal too small. Minimum: %dx%d. Current: %dx%d.",
			MinWidth, MinHeight, a.width, a.height)
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, msg)
	}

	// Title bar
	titleBar := a.renderTitleBar()

	// Tab bar
	tabBar := a.renderTabBar()

	// Separator
	sep := theme.Separator.Render(strings.Repeat("─", a.width))

	// Status bar
	statusBar := renderStatusBar(a.width, a.activeTab, "")

	// Content area height
	chromeHeight := lipgloss.Height(titleBar) + lipgloss.Height(tabBar) +
		lipgloss.Height(sep) + lipgloss.Height(statusBar)
	contentHeight := a.height - chromeHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Tab content
	content := a.activeTabView()

	// Truncate or pad content to fit
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}
	content = strings.Join(contentLines, "\n")

	// Compose
	screen := titleBar + "\n" + tabBar + "\n" + sep + "\n" + content + "\n" + statusBar

	// Help overlay
	if a.showHelp {
		helpView := a.help.View(a.activeTab)
		// Center the overlay
		helpWidth := lipgloss.Width(helpView)
		helpHeight := lipgloss.Height(helpView)
		x := (a.width - helpWidth) / 2
		y := (a.height - helpHeight) / 2
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		screen = overlayOnScreen(screen, helpView, x, y)
	}

	return screen
}

func (a App) renderTitleBar() string {
	left := " Mind Framework"
	if a.projectName != "" {
		left += " ─── " + a.projectName
	}
	if a.gitBranch != "" {
		left += " ── " + a.gitBranch
	}

	right := ""
	if a.version != "" {
		right = a.version + " "
	}

	gap := a.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return theme.TitleBar.Width(a.width).Render(
		left + strings.Repeat(" ", gap) + right)
}

func (a App) renderTabBar() string {
	var tabs []string
	for i, name := range TabNames {
		if TabID(i) == a.activeTab {
			tabs = append(tabs, theme.TabActive.Render("["+name+"]"))
		} else {
			tabs = append(tabs, theme.TabInactive.Render("["+name+"]"))
		}
	}
	return "  " + strings.Join(tabs, "  ")
}

func (a App) activeTabView() string {
	switch a.activeTab {
	case TabStatus:
		return a.status.View()
	case TabDocs:
		return a.docs.View()
	case TabIterations:
		return a.iters.View()
	case TabChecks:
		return a.checks.View()
	case TabQuality:
		return a.quality.View()
	}
	return ""
}

// Data loading commands

func (a App) loadHealth() tea.Cmd {
	return func() tea.Msg {
		project, err := fs.DetectProject(a.deps.ProjectRoot)
		if err != nil {
			return healthErrorMsg{err: err}
		}
		health, err := a.deps.ProjectSvc.AssembleHealth(project)
		if err != nil {
			return healthErrorMsg{err: err}
		}
		// Attach staleness
		staleness, err := a.deps.ReconcileSvc.ReadStaleness(a.deps.ProjectRoot)
		if err == nil && staleness != nil {
			health.Staleness = staleness
		}
		return healthLoadedMsg{health: health}
	}
}

func (a App) loadIterations() tea.Cmd {
	return func() tea.Msg {
		iters, err := a.deps.IterRepo.List()
		if err != nil {
			return iterationsErrorMsg{err: err}
		}
		return iterationsLoadedMsg{iterations: iters}
	}
}

func (a App) loadQuality() tea.Cmd {
	return func() tea.Msg {
		entries, err := a.deps.QualityRepo.ReadLog()
		if err != nil {
			return qualityErrorMsg{err: err}
		}
		return qualityLoadedMsg{entries: entries}
	}
}

// detectBranch reads the current git branch (best-effort).
func detectBranch(root string) string {
	// Simple: read .git/HEAD
	data, err := readFile(root + "/.git/HEAD")
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(data))
	if strings.HasPrefix(s, "ref: refs/heads/") {
		return strings.TrimPrefix(s, "ref: refs/heads/")
	}
	if len(s) >= 7 {
		return s[:7] // detached HEAD
	}
	return ""
}

// overlayOnScreen places overlay text on top of background at position (x, y).
func overlayOnScreen(bg, overlay string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	ovLines := strings.Split(overlay, "\n")

	for i, ovLine := range ovLines {
		bgRow := y + i
		if bgRow < 0 || bgRow >= len(bgLines) {
			continue
		}
		bgLine := bgLines[bgRow]
		// Replace characters at position x
		bgRunes := []rune(bgLine)
		ovRunes := []rune(ovLine)

		for len(bgRunes) < x+len(ovRunes) {
			bgRunes = append(bgRunes, ' ')
		}
		copy(bgRunes[x:], ovRunes)
		bgLines[bgRow] = string(bgRunes)
	}

	return strings.Join(bgLines, "\n")
}
