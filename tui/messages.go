package tui

import "github.com/jf-ferraz/mind-cli/domain"

// Data loading messages carry async results to tab models.

type healthLoadedMsg struct {
	health *domain.ProjectHealth
}

type healthErrorMsg struct {
	err error
}

type iterationsLoadedMsg struct {
	iterations []domain.Iteration
}

type iterationsErrorMsg struct {
	err error
}

type qualityLoadedMsg struct {
	entries []domain.QualityEntry
}

type qualityErrorMsg struct {
	err error
}

type validationCompleteMsg struct {
	report domain.UnifiedValidationReport
}

type validationErrorMsg struct {
	err error
}

type previewLoadedMsg struct {
	content string
}

type previewErrorMsg struct {
	err error
}

// UI state messages.

type validationStartedMsg struct{}

// tabActivatedMsg is sent when a tab becomes active for the first time.
type tabActivatedMsg struct {
	tab TabID
}
