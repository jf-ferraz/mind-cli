package mem

import (
	"fmt"
	"sort"

	"github.com/jf-ferraz/mind-cli/domain"
)

// IterationRepo is an in-memory implementation of repo.IterationRepo for testing.
type IterationRepo struct {
	Iterations []domain.Iteration
}

// NewIterationRepo creates an in-memory IterationRepo.
func NewIterationRepo() *IterationRepo {
	return &IterationRepo{}
}

// List returns all iterations, newest first.
func (r *IterationRepo) List() ([]domain.Iteration, error) {
	result := make([]domain.Iteration, len(r.Iterations))
	copy(result, r.Iterations)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Seq > result[j].Seq
	})
	return result, nil
}

// NextSeq returns the next available sequence number.
func (r *IterationRepo) NextSeq() (int, error) {
	if len(r.Iterations) == 0 {
		return 1, nil
	}
	max := 0
	for _, iter := range r.Iterations {
		if iter.Seq > max {
			max = iter.Seq
		}
	}
	return max + 1, nil
}

// Create creates a new iteration in memory.
func (r *IterationRepo) Create(reqType domain.RequestType, descriptor string) (*domain.Iteration, error) {
	seq, _ := r.NextSeq()
	dirName := fmt.Sprintf("%03d-%s-%s", seq, string(reqType), descriptor)
	iter := domain.Iteration{
		Seq:        seq,
		Type:       reqType,
		Descriptor: descriptor,
		DirName:    dirName,
		Status:     domain.IterInProgress,
	}
	for _, name := range domain.ExpectedArtifacts {
		exists := name == "overview.md"
		iter.Artifacts = append(iter.Artifacts, domain.Artifact{
			Name:   name,
			Exists: exists,
		})
	}
	r.Iterations = append(r.Iterations, iter)
	return &iter, nil
}
