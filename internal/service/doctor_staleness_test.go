package service

import (
	"strings"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// FR-81: Doctor reports stale documents as WARN diagnostics.
func TestDoctorService_FR81_StaleDiagnostics(t *testing.T) {
	docRepo := mem.NewDocRepo()
	docRepo.Dirs[".mind"] = true
	docRepo.Dirs["docs"] = true
	for _, zone := range domain.AllZones {
		docRepo.Dirs["docs/"+string(zone)] = true
	}
	docRepo.Files[".claude/CLAUDE.md"] = []byte("# Claude")
	docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\n## Vision\nOur vision\n\n## Key Deliverables\nStuff\n\n## Scope\nThings")
	docRepo.Files["docs/state/current.md"] = []byte("# Current\nContent")
	docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow\nContent")
	docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Index\nContent")
	docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary\nContent")

	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	briefRepo.Brief = &domain.Brief{
		GateResult: domain.BriefPresent,
	}
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
		Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
	}

	lockRepo := mem.NewLockRepo()
	lockRepo.Lock = &domain.LockFile{
		GeneratedAt: time.Now().UTC(),
		Status:      domain.LockStale,
		Entries: map[string]domain.LockEntry{
			"doc:spec/architecture": {
				ID:          "doc:spec/architecture",
				Stale:       true,
				StaleReason: "dependency changed: doc:spec/requirements (prerequisite changed)",
			},
			"doc:spec/domain-model": {
				ID:          "doc:spec/domain-model",
				Stale:       true,
				StaleReason: "dependency changed: doc:spec/requirements (may be outdated)",
			},
			"doc:spec/requirements": {
				ID: "doc:spec/requirements",
			},
		},
	}

	svc := NewDoctorService("/test", docRepo, iterRepo, briefRepo, configRepo, lockRepo)
	report := svc.Run(false)

	// Find staleness diagnostics
	staleCount := 0
	for _, d := range report.Diagnostics {
		if d.Category == "staleness" {
			staleCount++
			// Each should be WARN level
			if d.Level != domain.LevelWarn {
				t.Errorf("staleness diagnostic level = %q, want WARN", d.Level)
			}
			// Should have remediation text
			if !strings.Contains(d.Fix, "mind reconcile --force") {
				t.Errorf("staleness fix = %q, want to contain 'mind reconcile --force'", d.Fix)
			}
		}
	}

	if staleCount != 2 {
		t.Errorf("staleness diagnostics = %d, want 2", staleCount)
	}
}

// FR-81: Doctor with no lock file produces no staleness diagnostics.
func TestDoctorService_FR81_NoLock(t *testing.T) {
	docRepo := mem.NewDocRepo()
	docRepo.Dirs[".mind"] = true
	docRepo.Dirs["docs"] = true
	for _, zone := range domain.AllZones {
		docRepo.Dirs["docs/"+string(zone)] = true
	}
	docRepo.Files[".claude/CLAUDE.md"] = []byte("# Claude")
	docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\nContent")
	docRepo.Files["docs/state/current.md"] = []byte("# Current\nContent")
	docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow\nContent")
	docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Index\nContent")
	docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary\nContent")

	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	briefRepo.Brief = &domain.Brief{GateResult: domain.BriefPresent}
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
		Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
	}

	lockRepo := mem.NewLockRepo()
	// No lock file

	svc := NewDoctorService("/test", docRepo, iterRepo, briefRepo, configRepo, lockRepo)
	report := svc.Run(false)

	for _, d := range report.Diagnostics {
		if d.Category == "staleness" {
			t.Error("should not have staleness diagnostics when no lock file exists")
		}
	}
}

// FR-81: Doctor with nil lockRepo produces no staleness diagnostics.
func TestDoctorService_FR81_NilLockRepo(t *testing.T) {
	docRepo := mem.NewDocRepo()
	docRepo.Dirs[".mind"] = true
	docRepo.Dirs["docs"] = true

	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	briefRepo.Brief = &domain.Brief{GateResult: domain.BriefPresent}
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
		Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
	}

	svc := NewDoctorService("/test", docRepo, iterRepo, briefRepo, configRepo, nil)
	report := svc.Run(false)

	// Should not panic or error
	for _, d := range report.Diagnostics {
		if d.Category == "staleness" {
			t.Error("should not have staleness diagnostics when lockRepo is nil")
		}
	}
}

// FR-81: Doctor with CLEAN lock produces no staleness diagnostics.
func TestDoctorService_FR81_CleanLock(t *testing.T) {
	docRepo := mem.NewDocRepo()
	docRepo.Dirs[".mind"] = true
	docRepo.Dirs["docs"] = true
	for _, zone := range domain.AllZones {
		docRepo.Dirs["docs/"+string(zone)] = true
	}
	docRepo.Files[".claude/CLAUDE.md"] = []byte("# Claude")
	docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\nContent")
	docRepo.Files["docs/state/current.md"] = []byte("# Current\nContent")
	docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow\nContent")
	docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Index\nContent")
	docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary\nContent")

	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	briefRepo.Brief = &domain.Brief{GateResult: domain.BriefPresent}
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
		Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
	}

	lockRepo := mem.NewLockRepo()
	lockRepo.Lock = &domain.LockFile{
		Status: domain.LockClean,
		Entries: map[string]domain.LockEntry{
			"doc:spec/a": {ID: "doc:spec/a", Stale: false},
			"doc:spec/b": {ID: "doc:spec/b", Stale: false},
		},
	}

	svc := NewDoctorService("/test", docRepo, iterRepo, briefRepo, configRepo, lockRepo)
	report := svc.Run(false)

	for _, d := range report.Diagnostics {
		if d.Category == "staleness" {
			t.Error("should not have staleness diagnostics when lock is CLEAN")
		}
	}
}
