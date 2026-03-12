package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// QualityService reads and writes quality log entries.
type QualityService struct {
	projectRoot string
	qualityRepo repo.QualityRepo
}

// NewQualityService creates a QualityService.
func NewQualityService(projectRoot string, qualityRepo repo.QualityRepo) *QualityService {
	return &QualityService{projectRoot: projectRoot, qualityRepo: qualityRepo}
}

// ReadLog returns all quality entries.
func (s *QualityService) ReadLog() ([]domain.QualityEntry, error) {
	return s.qualityRepo.ReadLog()
}

// Log extracts quality scores from a convergence file and appends them to the quality log.
// filePath is relative to projectRoot. topic and variant override values found in the file.
func (s *QualityService) Log(filePath, topic, variant string) (*domain.QualityEntry, error) {
	absPath := filepath.Join(s.projectRoot, filePath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read convergence file: %w", err)
	}

	entry, err := parseConvergenceEntry(string(data), filePath)
	if err != nil {
		return nil, fmt.Errorf("parse convergence scores: %w", err)
	}

	if topic != "" {
		entry.Topic = topic
	}
	if variant != "" {
		entry.Variant = variant
	}
	if entry.Topic == "" {
		entry.Topic = strings.TrimSuffix(filepath.Base(filePath), ".md")
	}
	entry.OutputPath = filePath
	entry.Date = time.Now()

	if err := entry.Validate(); err != nil {
		return nil, fmt.Errorf("invalid quality entry: %w", err)
	}

	if err := s.appendEntry(entry); err != nil {
		return nil, fmt.Errorf("append quality entry: %w", err)
	}

	return entry, nil
}

// scoreRe matches "| dimension name | N |" table rows (supports underscores in dimension names).
var scoreRe = regexp.MustCompile(`(?i)\|\s*([\w]+(?:_[\w]+)*)\s*\|\s*(\d)\s*\|`)

// overallRe matches "Overall: N.N" or "Score: N.N"
var overallRe = regexp.MustCompile(`(?i)(?:overall|score)[:\s]+([0-9]+(?:\.[0-9]+)?)`)

func parseConvergenceEntry(content, path string) (*domain.QualityEntry, error) {
	entry := &domain.QualityEntry{}

	dimNames := []string{
		domain.DimPerspectiveDiversity,
		domain.DimEvidenceQuality,
		domain.DimConcessionDepth,
		domain.DimChallengeSubstantiveness,
		domain.DimSynthesisQuality,
		domain.DimActionability,
	}
	dimMap := map[string]bool{}
	for _, d := range dimNames {
		dimMap[d] = true
	}

	for _, m := range scoreRe.FindAllStringSubmatch(content, -1) {
		name := strings.ToLower(m[1])
		val, _ := strconv.Atoi(m[2])
		if dimMap[name] {
			entry.Dimensions = append(entry.Dimensions, domain.QualityDimension{Name: name, Value: val})
		}
	}

	// Fill missing dimensions with 0 to ensure exactly 6 dimensions are present
	found := map[string]bool{}
	for _, d := range entry.Dimensions {
		found[d.Name] = true
	}
	for _, name := range dimNames {
		if !found[name] {
			entry.Dimensions = append(entry.Dimensions, domain.QualityDimension{Name: name, Value: 0})
		}
	}

	// Overall score
	if m := overallRe.FindStringSubmatch(content); m != nil {
		entry.Score, _ = strconv.ParseFloat(m[1], 64)
	} else {
		// Average dimensions
		total := 0
		for _, d := range entry.Dimensions {
			total += d.Value
		}
		if len(entry.Dimensions) > 0 {
			entry.Score = float64(total) / float64(len(entry.Dimensions))
		}
	}

	entry.GatePass = entry.Score >= 3.0
	return entry, nil
}

func (s *QualityService) appendEntry(entry *domain.QualityEntry) error {
	logPath := filepath.Join(s.projectRoot, "docs", "knowledge", "quality-log.yml")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}

	line := fmt.Sprintf(
		"- topic: %s\n  variant: %s\n  date: %s\n  score: %.2f\n  gate_pass: %v\n  output_path: %s\n  dimensions:\n",
		entry.Topic,
		entry.Variant,
		entry.Date.Format("2006-01-02"),
		entry.Score,
		entry.GatePass,
		entry.OutputPath,
	)
	for _, d := range entry.Dimensions {
		line += fmt.Sprintf("    - name: %s\n      value: %d\n", d.Name, d.Value)
	}
	line += "\n"

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}
