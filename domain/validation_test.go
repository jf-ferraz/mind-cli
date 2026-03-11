package domain

import "testing"

// TestValidationReportOk verifies ValidationReport.Ok behavior.
func TestValidationReportOk(t *testing.T) {
	tests := []struct {
		name   string
		report ValidationReport
		want   bool
	}{
		{
			name:   "no checks is ok",
			report: ValidationReport{},
			want:   true,
		},
		{
			name: "all passed is ok",
			report: ValidationReport{
				Total:  3,
				Passed: 3,
			},
			want: true,
		},
		{
			name: "warnings only is ok",
			report: ValidationReport{
				Total:    3,
				Passed:   2,
				Warnings: 1,
			},
			want: true,
		},
		{
			name: "failures means not ok",
			report: ValidationReport{
				Total:  3,
				Passed: 2,
				Failed: 1,
			},
			want: false,
		},
		{
			name: "failures and warnings means not ok",
			report: ValidationReport{
				Total:    3,
				Passed:   1,
				Failed:   1,
				Warnings: 1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.report.Ok()
			if got != tt.want {
				t.Errorf("ValidationReport.Ok() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckLevelConstants verifies check level string values.
func TestCheckLevelConstants(t *testing.T) {
	if LevelFail != "FAIL" {
		t.Errorf("LevelFail = %q, want FAIL", LevelFail)
	}
	if LevelWarn != "WARN" {
		t.Errorf("LevelWarn = %q, want WARN", LevelWarn)
	}
	if LevelInfo != "INFO" {
		t.Errorf("LevelInfo = %q, want INFO", LevelInfo)
	}
}
