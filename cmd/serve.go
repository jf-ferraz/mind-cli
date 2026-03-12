package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/internal/deps"
	"github.com/jf-ferraz/mind-cli/internal/mcp"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server (JSON-RPC 2.0 over stdio)",
	Long: `mind serve starts a Model Context Protocol server on stdio.
Claude Code connects to it via .mcp.json and can call 16 tools:
mind_status, mind_doctor, mind_check_brief, mind_validate_docs,
mind_validate_refs, mind_list_iterations, mind_show_iteration,
mind_read_state, mind_update_state, mind_create_iteration,
mind_list_stubs, mind_check_gate, mind_log_quality, mind_search_docs,
mind_read_config, mind_suggest_next`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Resolve project root independently (serve bypasses PersistentPreRunE)
	root, err := resolveRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mind serve: project not found: %v\n", err)
		root, _ = os.Getwd()
	}

	d := deps.Build(root, render.New(render.ModePlain, 80))
	transport := mcp.NewStdioTransport()
	server := mcp.NewServer(transport, d)
	return server.Run()
}
