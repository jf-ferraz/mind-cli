package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/jf-ferraz/mind-cli/internal/globalconfig"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global mind configuration",
	Long:  "View, edit, and validate the global configuration at ~/.config/mind/config.toml.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current global configuration",
	RunE:  runConfigShow,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open global config in $EDITOR",
	RunE:  runConfigEdit,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	RunE:  runConfigPath,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate global configuration",
	RunE:  runConfigValidate,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configValidateCmd)
	rootCmd.AddCommand(configCmd)
}

func globalConfigRepo() *globalconfig.GlobalConfigRepo {
	return globalconfig.NewGlobalConfigRepo("")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	if flagJSON {
		cfg, err := repo.Read()
		if err != nil {
			return exitRuntime(err)
		}
		out, merr := json.MarshalIndent(cfg, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal config: %w", merr))
		}
		fmt.Println(string(out))
		return nil
	}
	text, err := repo.ShowConfig()
	if err != nil {
		return exitRuntime(err)
	}
	fmt.Print(text)
	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	if err := repo.EnsureExists(); err != nil {
		return exitRuntime(err)
	}
	editor, err := repo.EditorCommand()
	if err != nil {
		return exitRuntime(err)
	}
	c := exec.Command(editor, repo.ConfigPath())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return exitRuntime(fmt.Errorf("editor exited with error: %w", err))
	}
	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	path := repo.ConfigPath()
	if flagJSON {
		out, _ := json.MarshalIndent(map[string]string{"path": path}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println(path)
	}
	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	repo := globalConfigRepo()
	err := repo.ValidateConfig()
	if flagJSON {
		valid := err == nil
		msg := "ok"
		if !valid {
			msg = err.Error()
		}
		out, _ := json.MarshalIndent(map[string]interface{}{
			"valid":   valid,
			"message": msg,
		}, "", "  ")
		fmt.Println(string(out))
		if !valid {
			return exitQuiet(1)
		}
		return nil
	}
	if err != nil {
		fmt.Printf("Config invalid: %s\n", err)
		return exitQuiet(1)
	}
	fmt.Println("Config valid.")
	return nil
}
