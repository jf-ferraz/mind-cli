package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

// IterationRepo implements repo.IterationRepo using the filesystem.
type IterationRepo struct {
	projectRoot string
	iterDir     string
}

// NewIterationRepo creates an IterationRepo.
func NewIterationRepo(projectRoot string) *IterationRepo {
	return &IterationRepo{
		projectRoot: projectRoot,
		iterDir:     filepath.Join(projectRoot, "docs", "iterations"),
	}
}

var iterDirRe = regexp.MustCompile(`^(\d{3})-([A-Z_]+)-(.+)$`)

// List returns all iterations sorted by sequence number descending (newest first).
func (r *IterationRepo) List() ([]domain.Iteration, error) {
	entries, err := os.ReadDir(r.iterDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var iterations []domain.Iteration
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		matches := iterDirRe.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		seq, _ := strconv.Atoi(matches[1])
		iterPath := filepath.Join(r.iterDir, entry.Name())

		iter := domain.Iteration{
			Seq:        seq,
			Type:       domain.RequestType(matches[2]),
			Descriptor: matches[3],
			DirName:    entry.Name(),
			Path:       iterPath,
		}

		// Check artifacts
		for _, name := range domain.ExpectedArtifacts {
			artPath := filepath.Join(iterPath, name)
			_, artErr := os.Stat(artPath)
			iter.Artifacts = append(iter.Artifacts, domain.Artifact{
				Name:   name,
				Path:   artPath,
				Exists: artErr == nil,
			})
		}

		// Determine status
		allExist := true
		anyExist := false
		for _, a := range iter.Artifacts {
			if a.Exists {
				anyExist = true
			} else {
				allExist = false
			}
		}
		switch {
		case allExist:
			iter.Status = domain.IterComplete
		case anyExist:
			iter.Status = domain.IterIncomplete
		default:
			iter.Status = domain.IterInProgress
		}

		// Get creation time from overview.md or directory
		if info, err := entry.Info(); err == nil {
			iter.CreatedAt = info.ModTime()
		}

		iterations = append(iterations, iter)
	}

	// Sort newest first
	sort.Slice(iterations, func(i, j int) bool {
		return iterations[i].Seq > iterations[j].Seq
	})

	return iterations, nil
}

// NextSeq returns the next available sequence number.
func (r *IterationRepo) NextSeq() (int, error) {
	iterations, err := r.List()
	if err != nil {
		return 1, err
	}
	if len(iterations) == 0 {
		return 1, nil
	}
	return iterations[0].Seq + 1, nil
}

// Create creates a new iteration folder with template files.
func (r *IterationRepo) Create(reqType domain.RequestType, descriptor string) (*domain.Iteration, error) {
	seq, err := r.NextSeq()
	if err != nil {
		return nil, err
	}

	dirName := fmt.Sprintf("%03d-%s-%s", seq, string(reqType), descriptor)
	iterPath := filepath.Join(r.iterDir, dirName)

	if err := os.MkdirAll(iterPath, 0755); err != nil {
		return nil, err
	}

	// Create minimal overview.md
	overview := fmt.Sprintf("# %s\n\n- **Type**: %s\n- **Created**: %s\n\n## Scope\n\n## Requirement Traceability\n\n| Req ID | Description | Analyst | Developer | Reviewer |\n|--------|-------------|---------|-----------|----------|\n",
		strings.ReplaceAll(descriptor, "-", " "),
		string(reqType),
		"<!-- date -->",
	)
	if err := os.WriteFile(filepath.Join(iterPath, "overview.md"), []byte(overview), 0644); err != nil {
		return nil, err
	}

	iter := &domain.Iteration{
		Seq:        seq,
		Type:       reqType,
		Descriptor: descriptor,
		DirName:    dirName,
		Path:       iterPath,
		Status:     domain.IterInProgress,
	}

	return iter, nil
}
