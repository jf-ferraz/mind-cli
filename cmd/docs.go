package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/spf13/cobra"
)

var flagZone string

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Manage and inspect documentation",
}

var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documents by zone",
	RunE:  runDocsList,
}

var docsTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show documentation tree with stub annotations",
	RunE:  runDocsTree,
}

var docsStubsCmd = &cobra.Command{
	Use:   "stubs",
	Short: "List all stub documents",
	RunE:  runDocsStubs,
}

var docsSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search documents for a query string",
	Args:  cobra.ExactArgs(1),
	RunE:  runDocsSearch,
}

var docsOpenCmd = &cobra.Command{
	Use:   "open [path-or-id]",
	Short: "Open a document in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE:  runDocsOpen,
}

func init() {
	docsListCmd.Flags().StringVar(&flagZone, "zone", "", "Filter by zone (spec, blueprints, state, iterations, knowledge)")

	docsCmd.AddCommand(docsListCmd)
	docsCmd.AddCommand(docsTreeCmd)
	docsCmd.AddCommand(docsStubsCmd)
	docsCmd.AddCommand(docsSearchCmd)
	docsCmd.AddCommand(docsOpenCmd)
	rootCmd.AddCommand(docsCmd)
}

func runDocsList(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	docRepo := fs.NewDocRepo(root)

	// Validate zone flag if provided
	if flagZone != "" && !domain.ValidZone(flagZone) {
		fmt.Fprintf(os.Stderr, "Error: invalid zone %q. Valid zones: %s\n",
			flagZone, strings.Join(domain.ZoneNames(), ", "))
		os.Exit(1)
		return nil
	}

	var allDocs []domain.Document
	byZone := make(map[string]int)
	total := 0

	for _, zone := range domain.AllZones {
		docs, err := docRepo.ListByZone(zone)
		if err != nil {
			continue
		}
		byZone[string(zone)] = len(docs)
		total += len(docs)

		if flagZone == "" || string(zone) == flagZone {
			allDocs = append(allDocs, docs...)
		}
	}

	// Sort by zone order then alphabetically
	sort.Slice(allDocs, func(i, j int) bool {
		if allDocs[i].Zone != allDocs[j].Zone {
			return zoneOrder(allDocs[i].Zone) < zoneOrder(allDocs[j].Zone)
		}
		return allDocs[i].Path < allDocs[j].Path
	})

	list := &domain.DocumentList{
		Documents: allDocs,
		ByZone:    byZone,
		Total:     total,
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderDocumentList(list))
	return nil
}

func runDocsTree(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	docRepo := fs.NewDocRepo(root)
	docs, err := docRepo.ListAll()
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderDocTree(docs))
	return nil
}

func runDocsStubs(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	docRepo := fs.NewDocRepo(root)
	docs, err := docRepo.ListAll()
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	list := &domain.StubList{}
	for _, doc := range docs {
		if doc.IsStub {
			list.Stubs = append(list.Stubs, domain.StubEntry{
				Path: doc.Path,
				Zone: string(doc.Zone),
				Hint: fmt.Sprintf("Fill in content for %s", doc.Path),
			})
		}
	}
	list.Count = len(list.Stubs)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderStubList(list))

	if list.Count > 0 {
		os.Exit(1)
	}
	return nil
}

func runDocsSearch(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	query := strings.ToLower(args[0])
	results := &domain.SearchResults{Query: args[0]}

	docsDir := filepath.Join(root, "docs")
	err = filepath.WalkDir(docsDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		relPath, _ := filepath.Rel(root, path)
		var matches []domain.SearchMatch
		var lines []string

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			lines = append(lines, scanner.Text())
		}

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), query) {
				match := domain.SearchMatch{
					Line: i + 1,
					Text: line,
				}
				if i > 0 {
					match.ContextBefore = lines[i-1]
				}
				if i < len(lines)-1 {
					match.ContextAfter = lines[i+1]
				}
				matches = append(matches, match)
			}
		}

		if len(matches) > 0 {
			results.Results = append(results.Results, domain.SearchFileResult{
				Path:    relPath,
				Matches: matches,
			})
			results.TotalMatches += len(matches)
			results.FilesMatched++
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("search docs: %w", err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderSearchResults(results))
	return nil
}

func runDocsOpen(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	query := args[0]

	// Try direct path first
	if _, err := os.Stat(filepath.Join(root, query)); err == nil {
		return openInEditor(root, query)
	}

	// Try doc:zone/name format
	if strings.HasPrefix(query, "doc:") {
		parts := strings.SplitN(query[4:], "/", 2)
		if len(parts) == 2 {
			zonePath := filepath.Join("docs", parts[0], parts[1]+".md")
			if _, err := os.Stat(filepath.Join(root, zonePath)); err == nil {
				return openInEditor(root, zonePath)
			}
		}
	}

	// Fuzzy match
	docRepo := fs.NewDocRepo(root)
	docs, err := docRepo.ListAll()
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	queryLower := strings.ToLower(query)
	var matches []domain.Document
	for _, doc := range docs {
		nameLower := strings.ToLower(doc.Name)
		pathLower := strings.ToLower(doc.Path)
		if strings.Contains(nameLower, queryLower) || strings.Contains(pathLower, queryLower) {
			matches = append(matches, doc)
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no document matching %q found\n", query)
		os.Exit(1)
		return nil
	}

	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "Error: ambiguous match for %q. Matches:\n", query)
		for _, m := range matches {
			fmt.Fprintf(os.Stderr, "  %s\n", m.Path)
		}
		os.Exit(1)
		return nil
	}

	if flagJSON {
		mode := render.DetectMode(flagJSON, flagNoColor)
		r := render.New(mode, render.TermWidth())
		result := map[string]string{
			"path":     matches[0].Path,
			"abs_path": matches[0].AbsPath,
		}
		fmt.Print(r.RenderWorkflowStatus(nil)) // placeholder
		_ = result
		data := fmt.Sprintf(`{"path": %q, "abs_path": %q}`, matches[0].Path, matches[0].AbsPath)
		fmt.Println(data)
		return nil
	}

	return openInEditor(root, matches[0].Path)
}

func openInEditor(root, relPath string) error {
	if flagJSON {
		absPath := filepath.Join(root, relPath)
		fmt.Printf(`{"path": %q, "abs_path": %q}`+"\n", relPath, absPath)
		return nil
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Fprintln(os.Stderr, "Error: $EDITOR is not set")
		os.Exit(1)
		return nil
	}

	absPath := filepath.Join(root, relPath)
	c := exec.Command(editor, absPath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func zoneOrder(z domain.Zone) int {
	for i, zone := range domain.AllZones {
		if zone == z {
			return i
		}
	}
	return len(domain.AllZones)
}
