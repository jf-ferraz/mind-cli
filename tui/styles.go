package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds all Lip Gloss styles for the TUI.
type Theme struct {
	// Zone colors
	ZoneSpec       lipgloss.Style
	ZoneBlueprints lipgloss.Style
	ZoneState      lipgloss.Style
	ZoneIterations lipgloss.Style
	ZoneKnowledge  lipgloss.Style

	// Severity
	Pass lipgloss.Style
	Fail lipgloss.Style
	Warn lipgloss.Style
	Dim  lipgloss.Style

	// Chrome
	TitleBar    lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	StatusBar   lipgloss.Style
	Separator   lipgloss.Style

	// Content
	Selected lipgloss.Style
	Border   lipgloss.Style
	Heading  lipgloss.Style
	Subtle   lipgloss.Style

	// Progress bars
	BarFilled lipgloss.Style
	BarEmpty  lipgloss.Style

	// Iteration type colors
	TypeNew         lipgloss.Style
	TypeEnhancement lipgloss.Style
	TypeBugFix      lipgloss.Style
	TypeRefactor    lipgloss.Style
}

// DefaultTheme returns the standard TUI theme per BP-05 Section 6.
func DefaultTheme() Theme {
	return Theme{
		// Zone colors
		ZoneSpec:       lipgloss.NewStyle().Foreground(lipgloss.Color("#5f87ff")),
		ZoneBlueprints: lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd7ff")),
		ZoneState:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700")),
		ZoneIterations: lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd787")),
		ZoneKnowledge:  lipgloss.NewStyle().Foreground(lipgloss.Color("#d75fd7")),

		// Severity
		Pass: lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd787")),
		Fail: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")),
		Warn: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700")),
		Dim:  lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),

		// Chrome
		TitleBar:    lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1),
		TabActive:   lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("#ffffff")),
		TabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		StatusBar:   lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#aaaaaa")).Padding(0, 1),
		Separator:   lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),

		// Content
		Selected: lipgloss.NewStyle().Reverse(true),
		Border:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#444444")),
		Heading:  lipgloss.NewStyle().Bold(true),
		Subtle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),

		// Progress bars
		BarFilled: lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd787")),
		BarEmpty:  lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),

		// Iteration types
		TypeNew:         lipgloss.NewStyle().Foreground(lipgloss.Color("#5f87ff")),
		TypeEnhancement: lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd787")),
		TypeBugFix:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")),
		TypeRefactor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700")),
	}
}

var theme = DefaultTheme()
