package domain

import (
	"fmt"
	"time"
)

// QualityEntry represents a single convergence analysis result from quality-log.yml.
type QualityEntry struct {
	Topic      string             `json:"topic" yaml:"topic"`
	Variant    string             `json:"variant" yaml:"variant"`
	Date       time.Time          `json:"date" yaml:"date"`
	Score      float64            `json:"score" yaml:"score"`
	GatePass   bool               `json:"gate_pass" yaml:"gate_pass"`
	Dimensions []QualityDimension `json:"dimensions" yaml:"dimensions"`
	Personas   []string           `json:"personas" yaml:"personas"`
	OutputPath string             `json:"output_path" yaml:"output_path"`
}

// Validate checks BR-36, BR-37, BR-38.
func (e QualityEntry) Validate() error {
	if e.Score < 0.0 || e.Score > 5.0 {
		return fmt.Errorf("score %.2f outside valid range [0.0, 5.0]", e.Score)
	}
	if e.GatePass != (e.Score >= 3.0) {
		return fmt.Errorf("gate_pass=%v inconsistent with score %.2f", e.GatePass, e.Score)
	}
	if len(e.Dimensions) != 6 {
		return fmt.Errorf("expected 6 dimensions, got %d", len(e.Dimensions))
	}
	for _, d := range e.Dimensions {
		if d.Value < 0 || d.Value > 5 {
			return fmt.Errorf("dimension %s value %d outside valid range [0, 5]", d.Name, d.Value)
		}
	}
	return nil
}

// QualityDimension represents a single dimension score within a convergence analysis.
type QualityDimension struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

// Standard quality dimension names.
const (
	DimRigor         = "rigor"
	DimCoverage      = "coverage"
	DimActionability = "actionability"
	DimObjectivity   = "objectivity"
	DimConvergence   = "convergence"
	DimDepth         = "depth"
)
