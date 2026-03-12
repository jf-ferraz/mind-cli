package deps

import (
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
)

// Deps holds all repositories and services. Constructed once via Build,
// shared by both CLI command handlers and the TUI.
type Deps struct {
	ProjectRoot   string
	Renderer      *render.Renderer
	DocRepo       *fs.DocRepo
	IterRepo      *fs.IterationRepo
	BriefRepo     *fs.BriefRepo
	ConfigRepo    *fs.ConfigRepo
	LockRepo      *fs.LockRepo
	StateRepo     *fs.StateRepo
	QualityRepo   *fs.QualityRepo
	ProjectSvc    *service.ProjectService
	ValidationSvc *service.ValidationService
	ReconcileSvc  *service.ReconciliationService
	DoctorSvc     *service.DoctorService
	WorkflowSvc   *service.WorkflowService
	GenerateSvc   *service.GenerateService
}

// Build constructs all repositories and services for a given project root.
// Renderer may be nil when called from the TUI (which uses Lip Gloss directly).
func Build(root string, r *render.Renderer) *Deps {
	docRepo := fs.NewDocRepo(root)
	iterRepo := fs.NewIterationRepo(root)
	stateRepo := fs.NewStateRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)
	configRepo := fs.NewConfigRepo(root)
	lockRepo := fs.NewLockRepo(root)
	qualityRepo := fs.NewQualityRepo(root)

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
		ProjectSvc:    service.NewProjectService(docRepo, iterRepo, stateRepo, briefRepo),
		ValidationSvc: service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo),
		ReconcileSvc:  service.NewReconciliationService(configRepo, docRepo, lockRepo),
		DoctorSvc:     service.NewDoctorService(root, docRepo, iterRepo, briefRepo, configRepo, lockRepo),
		WorkflowSvc:   service.NewWorkflowService(stateRepo, iterRepo),
		GenerateSvc:   service.NewGenerateService(root),
	}
}
