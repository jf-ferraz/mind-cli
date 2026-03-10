package domain

import (
	"regexp"
	"strings"
	"time"
)

// RequestType classifies a user's workflow request.
type RequestType string

const (
	TypeNewProject  RequestType = "NEW_PROJECT"
	TypeBugFix      RequestType = "BUG_FIX"
	TypeEnhancement RequestType = "ENHANCEMENT"
	TypeRefactor    RequestType = "REFACTOR"
	TypeComplexNew  RequestType = "COMPLEX_NEW"
)

// IterationStatus represents the completeness of an iteration.
type IterationStatus string

const (
	IterInProgress IterationStatus = "in_progress"
	IterComplete   IterationStatus = "complete"
	IterIncomplete IterationStatus = "incomplete"
)

// Artifact represents a single file within an iteration folder.
type Artifact struct {
	Name   string // overview.md, changes.md, etc.
	Path   string // Absolute path
	Exists bool
}

// Iteration represents a single workflow iteration folder.
type Iteration struct {
	Seq        int             // Sequence number (1, 2, 3...)
	Type       RequestType     // NEW_PROJECT, BUG_FIX, etc.
	Descriptor string          // Kebab-case slug
	DirName    string          // Full directory name: "001-NEW_PROJECT-rest-api"
	Path       string          // Absolute path to iteration directory
	Artifacts  []Artifact      // Files in the iteration folder
	Status     IterationStatus // Derived from artifact presence
	CreatedAt  time.Time       // From overview.md mod time
}

// ExpectedArtifacts are the files expected in an iteration folder.
var ExpectedArtifacts = []string{
	"overview.md",
	"changes.md",
	"test-summary.md",
	"validation.md",
	"retrospective.md",
}

// slugRe matches non-alphanumeric characters for slugification.
var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a title to a kebab-case slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// Classify determines the RequestType from a natural language description.
func Classify(request string) RequestType {
	lower := strings.ToLower(request)

	// Explicit prefix matching (strongest signal)
	if strings.HasPrefix(lower, "analyze:") || strings.HasPrefix(lower, "explore:") {
		return TypeComplexNew
	}
	if strings.HasPrefix(lower, "create:") || strings.HasPrefix(lower, "build:") {
		return TypeNewProject
	}
	if strings.HasPrefix(lower, "fix:") {
		return TypeBugFix
	}
	if strings.HasPrefix(lower, "add:") {
		return TypeEnhancement
	}
	if strings.HasPrefix(lower, "refactor:") {
		return TypeRefactor
	}

	// Keyword matching
	bugKeywords := []string{"fix", "bug", "error", "broken", "crash", "regression", "failing"}
	for _, kw := range bugKeywords {
		if strings.Contains(lower, kw) {
			return TypeBugFix
		}
	}

	refactorKeywords := []string{"refactor", "clean", "restructure", "optimize", "simplify", "modernize"}
	for _, kw := range refactorKeywords {
		if strings.Contains(lower, kw) {
			return TypeRefactor
		}
	}

	enhanceKeywords := []string{"add", "feature", "extend", "improve", "integrate", "support"}
	for _, kw := range enhanceKeywords {
		if strings.Contains(lower, kw) {
			return TypeEnhancement
		}
	}

	newKeywords := []string{"create", "build", "new project", "scaffold"}
	for _, kw := range newKeywords {
		if strings.Contains(lower, kw) {
			return TypeNewProject
		}
	}

	// Default to enhancement for ambiguous requests
	return TypeEnhancement
}
