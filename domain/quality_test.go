package domain_test

import (
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
)

// validEntry returns a QualityEntry that passes all validation rules.
func validEntry() domain.QualityEntry {
	return domain.QualityEntry{
		Topic:    "auth-strategy",
		Variant:  "convergence-v1",
		Date:     time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Score:    4.0,
		GatePass: true,
		Dimensions: []domain.QualityDimension{
			{Name: domain.DimPerspectiveDiversity, Value: 4},
			{Name: domain.DimEvidenceQuality, Value: 4},
			{Name: domain.DimConcessionDepth, Value: 4},
			{Name: domain.DimChallengeSubstantiveness, Value: 4},
			{Name: domain.DimSynthesisQuality, Value: 4},
			{Name: domain.DimActionability, Value: 4},
		},
		Personas:   []string{"moderator", "analyst", "architect"},
		OutputPath: "docs/knowledge/auth-strategy-convergence.md",
	}
}

// BR-36: Score must be in [0.0, 5.0].
func TestQualityEntry_Validate_ScoreBounds(t *testing.T) {
	tests := []struct {
		name    string
		score   float64
		gate    bool
		wantErr bool
	}{
		{"zero score", 0.0, false, false},
		{"minimum passing", 3.0, true, false},
		{"maximum score", 5.0, true, false},
		{"mid-range score", 2.5, false, false},
		{"negative score", -0.1, false, true},
		{"above max", 5.1, true, true},
		{"way above max", 10.0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEntry()
			e.Score = tt.score
			e.GatePass = tt.gate
			err := e.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// BR-37: GatePass must equal (Score >= 3.0).
func TestQualityEntry_Validate_GateConsistency(t *testing.T) {
	tests := []struct {
		name    string
		score   float64
		gate    bool
		wantErr bool
	}{
		{"score 4.0 gate true", 4.0, true, false},
		{"score 4.0 gate false", 4.0, false, true},
		{"score 2.0 gate false", 2.0, false, false},
		{"score 2.0 gate true", 2.0, true, true},
		{"score 3.0 gate true", 3.0, true, false},
		{"score 3.0 gate false", 3.0, false, true},
		{"score 2.99 gate false", 2.99, false, false},
		{"score 2.99 gate true", 2.99, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEntry()
			e.Score = tt.score
			e.GatePass = tt.gate
			err := e.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// BR-38: Exactly 6 dimensions with values in [0, 5].
func TestQualityEntry_Validate_Dimensions(t *testing.T) {
	t.Run("5 dimensions", func(t *testing.T) {
		e := validEntry()
		e.Dimensions = e.Dimensions[:5]
		if err := e.Validate(); err == nil {
			t.Error("expected error for 5 dimensions")
		}
	})

	t.Run("7 dimensions", func(t *testing.T) {
		e := validEntry()
		e.Dimensions = append(e.Dimensions, domain.QualityDimension{Name: "extra", Value: 3})
		if err := e.Validate(); err == nil {
			t.Error("expected error for 7 dimensions")
		}
	})

	t.Run("0 dimensions", func(t *testing.T) {
		e := validEntry()
		e.Dimensions = nil
		if err := e.Validate(); err == nil {
			t.Error("expected error for 0 dimensions")
		}
	})

	t.Run("dimension value -1", func(t *testing.T) {
		e := validEntry()
		e.Dimensions[0].Value = -1
		if err := e.Validate(); err == nil {
			t.Error("expected error for negative dimension value")
		}
	})

	t.Run("dimension value 6", func(t *testing.T) {
		e := validEntry()
		e.Dimensions[0].Value = 6
		if err := e.Validate(); err == nil {
			t.Error("expected error for dimension value > 5")
		}
	})

	t.Run("dimension value 0", func(t *testing.T) {
		e := validEntry()
		e.Dimensions[0].Value = 0
		if err := e.Validate(); err != nil {
			t.Errorf("dimension value 0 should be valid, got: %v", err)
		}
	})

	t.Run("dimension value 5", func(t *testing.T) {
		e := validEntry()
		e.Dimensions[0].Value = 5
		if err := e.Validate(); err != nil {
			t.Errorf("dimension value 5 should be valid, got: %v", err)
		}
	})
}

// Valid entry should pass.
func TestQualityEntry_Validate_ValidEntry(t *testing.T) {
	e := validEntry()
	if err := e.Validate(); err != nil {
		t.Errorf("valid entry should pass: %v", err)
	}
}

// Dimension constants have correct values aligned with the conversation workflow rubric.
func TestQualityDimensionConstants(t *testing.T) {
	expected := map[string]string{
		"perspective_diversity":    domain.DimPerspectiveDiversity,
		"evidence_quality":         domain.DimEvidenceQuality,
		"concession_depth":         domain.DimConcessionDepth,
		"challenge_substantiveness": domain.DimChallengeSubstantiveness,
		"synthesis_quality":        domain.DimSynthesisQuality,
		"actionability":            domain.DimActionability,
	}
	for want, got := range expected {
		if got != want {
			t.Errorf("dimension constant %q = %q, want %q", got, got, want)
		}
	}
}
