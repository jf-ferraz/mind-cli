package deps

import (
	"github.com/jf-ferraz/mind-cli/internal/orchestrate"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
)

// Deps holds all repositories and services. Constructed once via Build,
// shared by both CLI command handlers and the TUI.
type Deps struct {
	ProjectRoot   string
	Renderer      *render.Renderer
	DocRepo       repo.DocRepo
	IterRepo      repo.IterationRepo
	BriefRepo     repo.BriefRepo
	ConfigRepo    repo.ConfigRepo
	LockRepo      repo.LockRepo
	StateRepo     repo.StateRepo
	QualityRepo   repo.QualityRepo
	ProjectSvc    *service.ProjectService
	ValidationSvc *service.ValidationService
	ReconcileSvc  *service.ReconciliationService
	DoctorSvc     *service.DoctorService
	WorkflowSvc   *service.WorkflowService
	GenerateSvc   *service.GenerateService
	QualitySvc    *service.QualityService
	HandoffSvc    *orchestrate.HandoffService
}

// BuildOpts configures how Deps is constructed.
type BuildOpts struct {
	// GlobalOnly skips project root detection — only global services available.
	GlobalOnly bool
}

// Build constructs all repositories and services for a given project root.
// Renderer may be nil when called from the TUI (which uses Lip Gloss directly).
func Build(root string, r *render.Renderer) *Deps {
	return BuildWithOpts(root, r, BuildOpts{})
}

// BuildWithOpts constructs Deps with configurable options.
// When opts.GlobalOnly is true, project-scoped repos and services are nil.
func BuildWithOpts(root string, r *render.Renderer, opts BuildOpts) *Deps {
	if opts.GlobalOnly {
		return &Deps{
			Renderer: r,
		}
	}
	docRepo := fs.NewDocRepo(root)
	iterRepo := fs.NewIterationRepo(root)
	stateRepo := fs.NewStateRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)
	configRepo := fs.NewConfigRepo(root)
	lockRepo := fs.NewLockRepo(root)
	qualityRepo := fs.NewQualityRepo(root)

	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	return &Deps{
		ProjectRoot:   root,
		Renderer:      r,
		DocRepo:       docRepo,
		IterRepo:      iterRepo,
		BriefRepo:     briefRepo,
		ConfigRepo:    configRepo,
		LockRepo:      lockRepo,
		StateRepo:     stateRepo,
		QualityRepo:   qualityRepo,
		ProjectSvc:    service.NewProjectServiceWithConfig(docRepo, iterRepo, stateRepo, briefRepo, configRepo),
		ValidationSvc: validationSvc,
		ReconcileSvc:  service.NewReconciliationService(configRepo, docRepo, lockRepo),
		DoctorSvc:     service.NewDoctorService(root, docRepo, iterRepo, briefRepo, configRepo, lockRepo),
		WorkflowSvc:   service.NewWorkflowService(stateRepo, iterRepo),
		GenerateSvc:   service.NewGenerateService(root),
		QualitySvc:    service.NewQualityService(root, qualityRepo),
		HandoffSvc:    orchestrate.NewHandoffService(root, iterRepo, stateRepo, validationSvc),
	}
}
