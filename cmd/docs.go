package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
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

	fmt.Print(renderer.RenderDocumentList(list))
	return nil
}

func runDocsTree(cmd *cobra.Command, args []string) error {
	docs, err := docRepo.ListAll()
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	fmt.Print(renderer.RenderDocTree(docs))
	return nil
}

func runDocsStubs(cmd *cobra.Command, args []string) error {
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

	fmt.Print(renderer.RenderStubList(list))

	if list.Count > 0 {
		os.Exit(1)
	}
	return nil
}

func runDocsSearch(cmd *cobra.Command, args []string) error {
	results, err := docRepo.Search(args[0])
	if err != nil {
		return fmt.Errorf("search docs: %w", err)
	}
	fmt.Print(renderer.RenderSearchResults(results))
	return nil
}

func runDocsOpen(cmd *cobra.Command, args []string) error {
	query := args[0]

	// Try direct path first
	if _, err := os.Stat(filepath.Join(projectRoot, query)); err == nil {
		return openInEditor(projectRoot, query)
	}

	// Try doc:zone/name format
	if strings.HasPrefix(query, "doc:") {
		parts := strings.SplitN(query[4:], "/", 2)
		if len(parts) == 2 {
			zonePath := filepath.Join("docs", parts[0], parts[1]+".md")
			if _, err := os.Stat(filepath.Join(projectRoot, zonePath)); err == nil {
				return openInEditor(projectRoot, zonePath)
			}
		}
	}

	// Fuzzy match
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

	return openInEditor(projectRoot, matches[0].Path)
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
