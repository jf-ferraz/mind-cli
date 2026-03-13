package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage the global project registry",
	Long:  "List, add, remove, resolve, and check projects registered in ~/.config/mind/projects.toml.",
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered projects",
	RunE:  runRegistryList,
}

var registryAddCmd = &cobra.Command{
	Use:   "add <alias> <path>",
	Short: "Register a project by alias and path",
	Args:  cobra.ExactArgs(2),
	RunE:  runRegistryAdd,
}

var registryRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove a project by alias",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryRemove,
}

var registryResolveCmd = &cobra.Command{
	Use:   "resolve <@alias>",
	Short: "Resolve an @alias to its absolute path",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryResolve,
}

var registryCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate all registered project paths exist",
	RunE:  runRegistryCheck,
}

func init() {
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryResolveCmd)
	registryCmd.AddCommand(registryCheckCmd)
	rootCmd.AddCommand(registryCmd)
}

func runRegistryList(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	items, err := repo.RegistryList()
	if err != nil {
		return exitRuntime(err)
	}
	if flagJSON {
		out, merr := json.MarshalIndent(items, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal registry list: %w", merr))
		}
		fmt.Println(string(out))
		return nil
	}
	if len(items) == 0 {
		fmt.Println("No projects registered.")
		fmt.Println("Run: mind registry add <alias> <path>")
		return nil
	}
	for _, item := range items {
		fmt.Printf("  %-20s %s\n", item.Alias, item.Path)
	}
	return nil
}

func runRegistryAdd(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	alias, path := args[0], args[1]
	if err := repo.RegistryAdd(alias, path); err != nil {
		return exitValidation(err)
	}
	if flagJSON {
		out, _ := json.MarshalIndent(map[string]string{
			"alias":  alias,
			"status": "added",
		}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Printf("Added %q -> %s\n", alias, path)
	}
	return nil
}

func runRegistryRemove(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	alias := args[0]
	if err := repo.RegistryRemove(alias); err != nil {
		return exitValidation(err)
	}
	if flagJSON {
		out, _ := json.MarshalIndent(map[string]string{
			"alias":  alias,
			"status": "removed",
		}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Printf("Removed %q\n", alias)
	}
	return nil
}

func runRegistryResolve(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	path, err := repo.RegistryResolve(args[0])
	if err != nil {
		return exitValidation(err)
	}
	if flagJSON {
		out, _ := json.MarshalIndent(map[string]string{
			"alias": args[0],
			"path":  path,
		}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println(path)
	}
	return nil
}

func runRegistryCheck(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	results, err := repo.RegistryCheck()
	if err != nil {
		return exitRuntime(err)
	}
	if flagJSON {
		out, merr := json.MarshalIndent(results, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal registry check: %w", merr))
		}
		fmt.Println(string(out))
		return nil
	}
	allOK := true
	for _, r := range results {
		if r.Exists {
			fmt.Printf("  OK %-20s %s\n", r.Alias, r.Path)
		} else {
			fmt.Printf("  !! %-20s %s\n", r.Alias, r.Error)
			allOK = false
		}
	}
	if !allOK {
		return exitQuiet(1)
	}
	return nil
}
