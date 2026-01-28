package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/cliq-cli/cliq/internal/config"
	"github.com/cliq-cli/cliq/internal/parser"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Cliq configuration",
	Long: `Manage Cliq configuration including viewing parsed configurations,
reloading config files, and editing the configuration.

Subcommands:
  show    Show parsed configuration
  reload  Reload and re-parse configs
  edit    Open config file in $EDITOR`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// showCmd represents the config show command
var showCmd = &cobra.Command{
	Use:   "show [nvim|tmux|all]",
	Short: "Show parsed configuration",
	Long:  `Display the parsed Neovim or tmux configuration, including detected keymaps and settings.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConfigShow,
}

// reloadCmd represents the config reload command
var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload and re-parse configs",
	Long:  `Re-parse all configuration files and update the cache.`,
	RunE:  runConfigReload,
}

// editCmd represents the config edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in $EDITOR",
	Long:  `Open the Cliq configuration file in your default editor.`,
	RunE:  runConfigEdit,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(reloadCmd)
	configCmd.AddCommand(editCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	target := "all"
	if len(args) > 0 {
		target = args[0]
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	switch target {
	case "nvim":
		return showNvimConfig(cfg, titleStyle, labelStyle)
	case "tmux":
		return showTmuxConfig(cfg, titleStyle, labelStyle)
	case "all":
		fmt.Println(titleStyle.Render("=== Cliq Configuration ===\n"))

		// Show general config
		fmt.Println(labelStyle.Render("Config File:"), config.GetConfigPath())
		fmt.Println(labelStyle.Render("Model Path:"), cfg.GetModelPath())
		fmt.Println(labelStyle.Render("Response Style:"), cfg.General.ResponseStyle)
		fmt.Println()

		if err := showNvimConfig(cfg, titleStyle, labelStyle); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		fmt.Println()

		if err := showTmuxConfig(cfg, titleStyle, labelStyle); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	default:
		return fmt.Errorf("unknown target: %s (use nvim, tmux, or all)", target)
	}

	return nil
}

func showNvimConfig(cfg *config.Config, titleStyle, labelStyle lipgloss.Style) error {
	fmt.Println(titleStyle.Render("--- Neovim Configuration ---"))

	if cfg.Nvim.ConfigPath == "" {
		fmt.Println("  No Neovim configuration detected")
		return nil
	}

	fmt.Println(labelStyle.Render("Config Path:"), cfg.Nvim.ConfigPath)

	nvimConfig, err := parser.ParseNvimConfig(cfg.Nvim.ConfigPath)
	if err != nil {
		return fmt.Errorf("could not parse nvim config: %w", err)
	}

	fmt.Println(labelStyle.Render("Leader Key:"), nvimConfig.Leader)
	fmt.Println(labelStyle.Render("Keymaps Found:"), len(nvimConfig.Keymaps))
	fmt.Println(labelStyle.Render("Plugins Found:"), len(nvimConfig.Plugins))

	if len(nvimConfig.Keymaps) > 0 {
		fmt.Println(labelStyle.Render("\nSample Keymaps:"))
		for i, km := range nvimConfig.Keymaps {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(nvimConfig.Keymaps)-5)
				break
			}
			fmt.Printf("  [%s] %s -> %s\n", km.Mode, km.Lhs, km.Rhs)
		}
	}

	if len(nvimConfig.Plugins) > 0 {
		fmt.Println(labelStyle.Render("\nDetected Plugins:"))
		for i, p := range nvimConfig.Plugins {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(nvimConfig.Plugins)-10)
				break
			}
			status := "enabled"
			if !p.Enabled {
				status = "disabled"
			}
			fmt.Printf("  %s (%s)\n", p.Name, status)
		}
	}

	return nil
}

func showTmuxConfig(cfg *config.Config, titleStyle, labelStyle lipgloss.Style) error {
	fmt.Println(titleStyle.Render("--- Tmux Configuration ---"))

	if cfg.Tmux.ConfigPath == "" {
		fmt.Println("  No tmux configuration detected")
		return nil
	}

	fmt.Println(labelStyle.Render("Config Path:"), cfg.Tmux.ConfigPath)

	tmuxConfig, err := parser.ParseTmuxConfig(cfg.Tmux.ConfigPath)
	if err != nil {
		return fmt.Errorf("could not parse tmux config: %w", err)
	}

	fmt.Println(labelStyle.Render("Prefix:"), tmuxConfig.Prefix)
	fmt.Println(labelStyle.Render("Keymaps Found:"), len(tmuxConfig.Keymaps))

	if len(tmuxConfig.Keymaps) > 0 {
		fmt.Println(labelStyle.Render("\nSample Keymaps:"))
		for i, km := range tmuxConfig.Keymaps {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(tmuxConfig.Keymaps)-5)
				break
			}
			fmt.Printf("  %s -> %s\n", km.Key, km.Command)
		}
	}

	return nil
}

func runConfigReload(cmd *cobra.Command, args []string) error {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	fmt.Println(titleStyle.Render("Reloading configurations..."))

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// Parse nvim config
	var nvimConfig *parser.NvimConfig
	if cfg.Nvim.ConfigPath != "" {
		nvimConfig, err = parser.ParseNvimConfig(cfg.Nvim.ConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse nvim config: %v\n", err)
		} else {
			fmt.Println(successStyle.Render("  ✓ Neovim config parsed"))
		}
	}

	// Parse tmux config
	var tmuxConfig *parser.TmuxConfig
	if cfg.Tmux.ConfigPath != "" {
		tmuxConfig, err = parser.ParseTmuxConfig(cfg.Tmux.ConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse tmux config: %v\n", err)
		} else {
			fmt.Println(successStyle.Render("  ✓ Tmux config parsed"))
		}
	}

	// Save cache
	cache := &parser.Cache{
		NvimConfig: nvimConfig,
		TmuxConfig: tmuxConfig,
	}

	if err := cache.Save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	fmt.Println(successStyle.Render("  ✓ Cache updated"))

	// Print summary as JSON for debugging
	if verbose {
		summary := map[string]interface{}{
			"nvim_keymaps":  len(nvimConfig.Keymaps),
			"nvim_plugins":  len(nvimConfig.Plugins),
			"tmux_keymaps":  len(tmuxConfig.Keymaps),
			"tmux_prefix":   tmuxConfig.Prefix,
		}
		data, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println("\nSummary:")
		fmt.Println(string(data))
	}

	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	configPath := config.GetConfigPath()

	// Check if config exists, create default if not
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := config.Default()
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	c := exec.Command(editor, configPath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
