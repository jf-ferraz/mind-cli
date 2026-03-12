package tui

// TabID identifies a dashboard tab.
type TabID int

const (
	TabStatus     TabID = iota // 0
	TabDocs                    // 1
	TabIterations              // 2
	TabChecks                  // 3
	TabQuality                 // 4
)

// TabCount is the total number of tabs.
const TabCount = 5

// TabNames maps TabID to display labels.
var TabNames = [TabCount]string{
	"1 Status",
	"2 Docs",
	"3 Iterations",
	"4 Check",
	"5 Quality",
}

// ViewState tracks data loading status for a tab.
type ViewState int

const (
	ViewLoading ViewState = iota
	ViewError
	ViewEmpty
	ViewReady
)

// MinWidth is the minimum terminal width.
const MinWidth = 80

// MinHeight is the minimum terminal height.
const MinHeight = 24
