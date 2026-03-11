package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/spf13/cobra"
)

var (
	flagReconcileCheck bool
	flagReconcileForce bool
	flagReconcileGraph bool
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Reconcile document hashes and propagate staleness",
	Long: `Scans all declared documents, computes SHA-256 hashes, detects changes
against the previous mind.lock state, and propagates staleness through
the dependency graph declared in mind.toml [[graph]].

Flags:
  --check   Read-only verification (exit 0 if clean, 4 if stale)
  --force   Discard lock, re-hash everything, clear staleness
  --graph   Show ASCII dependency graph visualization`,
	RunE: runReconcile,
}

func init() {
	reconcileCmd.Flags().BoolVar(&flagReconcileCheck, "check", false, "Read-only verification (exit 0 if clean, 4 if stale)")
	reconcileCmd.Flags().BoolVar(&flagReconcileForce, "force", false, "Discard lock, re-hash everything, clear staleness")
	reconcileCmd.Flags().BoolVar(&flagReconcileGraph, "graph", false, "Show ASCII dependency graph visualization")
	rootCmd.AddCommand(reconcileCmd)
}

func runReconcile(cmd *cobra.Command, args []string) error {
	opts := domain.ReconcileOpts{
		CheckOnly: flagReconcileCheck,
		Force:     flagReconcileForce,
		GraphOnly: flagReconcileGraph,
	}

	// --graph: render graph and exit (does not run full reconciliation)
	if opts.GraphOnly {
		return runReconcileGraph()
	}

	result, err := reconcileSvc.Reconcile(projectRoot, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if isConfigError(err) {
			os.Exit(3)
		}
		os.Exit(2)
		return nil
	}

	fmt.Print(renderer.RenderReconcileResult(result))

	// Exit code 4 for stale in --check mode (FR-82)
	if opts.CheckOnly && result.Status == domain.LockStale {
		os.Exit(4)
	}

	return nil
}

func runReconcileGraph() error {
	graph, stale, err := reconcileSvc.LoadGraph(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if isConfigError(err) {
			os.Exit(3)
		}
		os.Exit(2)
		return nil
	}

	fmt.Print(renderer.RenderGraph(graph, stale))
	return nil
}

// isConfigError returns true if the error is a configuration-related error.
func isConfigError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "mind.toml") || strings.Contains(msg, "no configuration")
}
