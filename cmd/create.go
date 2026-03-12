package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/generate"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create project artifacts (ADR, blueprint, iteration, spike, convergence, brief)",
}

var createADRCmd = &cobra.Command{
	Use:   "adr [title]",
	Short: "Create an auto-numbered ADR",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateADR,
}

var createBlueprintCmd = &cobra.Command{
	Use:   "blueprint [title]",
	Short: "Create an auto-numbered blueprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateBlueprint,
}

var createIterationCmd = &cobra.Command{
	Use:   "iteration [type] [name]",
	Short: "Create a new iteration (type: new, enhancement, bugfix, refactor)",
	Args:  cobra.ExactArgs(2),
	RunE:  runCreateIteration,
}

var createSpikeCmd = &cobra.Command{
	Use:   "spike [title]",
	Short: "Create a spike report template",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateSpike,
}

var createConvergenceCmd = &cobra.Command{
	Use:   "convergence [title]",
	Short: "Create a convergence analysis template",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateConvergence,
}

var createBriefCmd = &cobra.Command{
	Use:   "brief",
	Short: "Create project brief interactively",
	RunE:  runCreateBrief,
}

func init() {
	createCmd.AddCommand(createADRCmd)
	createCmd.AddCommand(createBlueprintCmd)
	createCmd.AddCommand(createIterationCmd)
	createCmd.AddCommand(createSpikeCmd)
	createCmd.AddCommand(createConvergenceCmd)
	createCmd.AddCommand(createBriefCmd)
	rootCmd.AddCommand(createCmd)
}

func runCreateADR(cmd *cobra.Command, args []string) error {
	result, err := generateSvc.CreateADR(args[0])
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return exitValidation(err)
		}
		return err
	}

	fmt.Print(renderer.RenderCreateResult(result))
	return nil
}

func runCreateBlueprint(cmd *cobra.Command, args []string) error {
	result, err := generateSvc.CreateBlueprint(args[0])
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return exitValidation(err)
		}
		return err
	}

	fmt.Print(renderer.RenderCreateResult(result))
	return nil
}

func runCreateIteration(cmd *cobra.Command, args []string) error {
	result, err := generateSvc.CreateIteration(args[0], args[1])
	if err != nil {
		return exitValidation(err)
	}

	fmt.Print(renderer.RenderCreateIterationResult(result))
	return nil
}

func runCreateSpike(cmd *cobra.Command, args []string) error {
	result, err := generateSvc.CreateSpike(args[0])
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return exitValidation(err)
		}
		return err
	}

	fmt.Print(renderer.RenderCreateResult(result))
	return nil
}

func runCreateConvergence(cmd *cobra.Command, args []string) error {
	result, err := generateSvc.CreateConvergence(args[0])
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return exitValidation(err)
		}
		return err
	}

	fmt.Print(renderer.RenderCreateResult(result))
	return nil
}

func runCreateBrief(cmd *cobra.Command, args []string) error {
	if flagJSON {
		return exitValidation(fmt.Errorf("'mind create brief' is interactive-only (--json not supported)"))
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return exitValidation(fmt.Errorf("'mind create brief' requires an interactive terminal. Edit docs/spec/project-brief.md directly."))
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Create Project Brief")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println()

	vision := promptMultiline(reader, "Vision (describe the project vision, empty line to finish):")
	deliverables := promptMultiline(reader, "Key Deliverables (list deliverables, empty line to finish):")
	inScope := promptMultiline(reader, "In Scope (what is in scope, empty line to finish):")
	outScope := promptMultiline(reader, "Out of Scope (what is out of scope, empty line to finish):")
	constraints := promptMultiline(reader, "Constraints (list constraints, empty line to finish):")

	content := generate.BriefTemplate(vision, deliverables, inScope, outScope, constraints)
	briefPath := fmt.Sprintf("%s/docs/spec/project-brief.md", projectRoot)

	if err := os.MkdirAll(fmt.Sprintf("%s/docs/spec", projectRoot), 0755); err != nil {
		return fmt.Errorf("create spec dir: %w", err)
	}

	if err := os.WriteFile(briefPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write brief: %w", err)
	}

	fmt.Printf("\nCreated: docs/spec/project-brief.md\n")
	return nil
}

func promptMultiline(reader *bufio.Reader, prompt string) string {
	fmt.Println(prompt)
	var lines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\n\r")
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
