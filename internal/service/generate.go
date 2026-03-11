package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/generate"
)

// GenerateService orchestrates document scaffolding and sequence derivation.
type GenerateService struct {
	projectRoot string
}

// NewGenerateService creates a GenerateService.
func NewGenerateService(projectRoot string) *GenerateService {
	return &GenerateService{projectRoot: projectRoot}
}

// CreateADR creates an auto-numbered ADR file.
func (s *GenerateService) CreateADR(title string) (*domain.CreateResult, error) {
	dir := filepath.Join(s.projectRoot, "docs", "spec", "decisions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create decisions dir: %w", err)
	}

	seq := s.nextSeq(dir, regexp.MustCompile(`^(\d+)-`))
	slug := domain.Slugify(title)
	filename := fmt.Sprintf("%03d-%s.md", seq, slug)
	absPath := filepath.Join(dir, filename)

	if _, err := os.Stat(absPath); err == nil {
		return nil, fmt.Errorf("create adr: %w: %s", domain.ErrAlreadyExists, filename)
	}

	content := generate.ADRTemplate(title, seq)
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write adr: %w", err)
	}

	relPath, _ := filepath.Rel(s.projectRoot, absPath)
	return &domain.CreateResult{
		Path:  relPath,
		Seq:   seq,
		Title: title,
	}, nil
}

// CreateBlueprint creates an auto-numbered blueprint file and updates INDEX.md.
func (s *GenerateService) CreateBlueprint(title string) (*domain.CreateResult, error) {
	dir := filepath.Join(s.projectRoot, "docs", "blueprints")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create blueprints dir: %w", err)
	}

	seq := s.nextSeq(dir, regexp.MustCompile(`^(\d+)-`))
	slug := domain.Slugify(title)
	filename := fmt.Sprintf("%02d-%s.md", seq, slug)
	absPath := filepath.Join(dir, filename)

	if _, err := os.Stat(absPath); err == nil {
		return nil, fmt.Errorf("create blueprint: %w: %s", domain.ErrAlreadyExists, filename)
	}

	content := generate.BlueprintTemplate(title, seq)
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write blueprint: %w", err)
	}

	// Update INDEX.md
	indexPath := filepath.Join(dir, "INDEX.md")
	indexUpdated := false
	entry := generate.IndexEntry(seq, slug, filename)

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		indexContent := generate.IndexStub() + "\n" + entry
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			return nil, fmt.Errorf("write INDEX.md: %w", err)
		}
		indexUpdated = true
	} else {
		f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("open INDEX.md: %w", err)
		}
		if _, err := f.WriteString(entry); err != nil {
			f.Close()
			return nil, fmt.Errorf("append to INDEX.md: %w", err)
		}
		f.Close()
		indexUpdated = true
	}

	relPath, _ := filepath.Rel(s.projectRoot, absPath)
	return &domain.CreateResult{
		Path:         relPath,
		Seq:          seq,
		Title:        title,
		IndexUpdated: indexUpdated,
	}, nil
}

// CreateIteration creates a new iteration directory with 5 template files.
func (s *GenerateService) CreateIteration(typeName string, name string) (*domain.CreateIterationResult, error) {
	typeMap := map[string]domain.RequestType{
		"new":         domain.TypeNewProject,
		"enhancement": domain.TypeEnhancement,
		"bugfix":      domain.TypeBugFix,
		"refactor":    domain.TypeRefactor,
	}

	reqType, ok := typeMap[typeName]
	if !ok {
		return nil, fmt.Errorf("invalid iteration type %q (expected: new, enhancement, bugfix, refactor)", typeName)
	}

	dir := filepath.Join(s.projectRoot, "docs", "iterations")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create iterations dir: %w", err)
	}

	seq := s.nextIterSeq(dir)
	slug := domain.Slugify(name)
	dirName := fmt.Sprintf("%03d-%s-%s", seq, string(reqType), slug)
	iterPath := filepath.Join(dir, dirName)

	if _, err := os.Stat(iterPath); err == nil {
		return nil, fmt.Errorf("create iteration: %w: %s", domain.ErrAlreadyExists, dirName)
	}

	if err := os.MkdirAll(iterPath, 0755); err != nil {
		return nil, fmt.Errorf("create iteration dir: %w", err)
	}

	templates := map[string]string{
		"overview.md":      generate.IterationOverviewTemplate(slug, string(reqType)),
		"changes.md":       generate.IterationChangesTemplate(),
		"test-summary.md":  generate.IterationTestSummaryTemplate(),
		"validation.md":    generate.IterationValidationTemplate(),
		"retrospective.md": generate.IterationRetrospectiveTemplate(),
	}

	var files []string
	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(iterPath, name), []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", name, err)
		}
		files = append(files, name)
	}
	sort.Strings(files)

	relPath, _ := filepath.Rel(s.projectRoot, iterPath)
	return &domain.CreateIterationResult{
		Path:       relPath,
		Seq:        seq,
		Type:       reqType,
		Descriptor: slug,
		Files:      files,
	}, nil
}

// CreateSpike creates a spike report template.
func (s *GenerateService) CreateSpike(title string) (*domain.CreateResult, error) {
	dir := filepath.Join(s.projectRoot, "docs", "knowledge")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create knowledge dir: %w", err)
	}

	slug := domain.Slugify(title)
	filename := slug + "-spike.md"
	absPath := filepath.Join(dir, filename)

	if _, err := os.Stat(absPath); err == nil {
		return nil, fmt.Errorf("create spike: %w: %s", domain.ErrAlreadyExists, filename)
	}

	content := generate.SpikeTemplate(title)
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write spike: %w", err)
	}

	relPath, _ := filepath.Rel(s.projectRoot, absPath)
	return &domain.CreateResult{
		Path:  relPath,
		Title: title,
	}, nil
}

// CreateConvergence creates a convergence analysis template.
func (s *GenerateService) CreateConvergence(title string) (*domain.CreateResult, error) {
	dir := filepath.Join(s.projectRoot, "docs", "knowledge")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create knowledge dir: %w", err)
	}

	slug := domain.Slugify(title)
	filename := slug + "-convergence.md"
	absPath := filepath.Join(dir, filename)

	if _, err := os.Stat(absPath); err == nil {
		return nil, fmt.Errorf("create convergence: %w: %s", domain.ErrAlreadyExists, filename)
	}

	content := generate.ConvergenceTemplate(title)
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write convergence: %w", err)
	}

	relPath, _ := filepath.Rel(s.projectRoot, absPath)
	return &domain.CreateResult{
		Path:  relPath,
		Title: title,
	}, nil
}

func (s *GenerateService) nextSeq(dir string, re *regexp.Regexp) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}
	max := 0
	for _, e := range entries {
		matches := re.FindStringSubmatch(e.Name())
		if matches == nil {
			continue
		}
		n, _ := strconv.Atoi(matches[1])
		if n > max {
			max = n
		}
	}
	return max + 1
}

var iterDirPattern = regexp.MustCompile(`^(\d{3})-`)

func (s *GenerateService) nextIterSeq(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		matches := iterDirPattern.FindStringSubmatch(e.Name())
		if matches == nil {
			continue
		}
		n, _ := strconv.Atoi(matches[1])
		if n > max {
			max = n
		}
	}
	return max + 1
}
