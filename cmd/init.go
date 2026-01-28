package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/cliq-cli/cliq/internal/config"
	"github.com/cliq-cli/cliq/internal/llm"
)

var (
	modelURL   string
	skipConfig bool
	forceInit  bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Cliq - downloads model and sets up configuration",
	Long: `Initialize Cliq by downloading the required language model and
setting up the configuration. This command will:

1. Create configuration directories
2. Download the Phi-3-mini model (~2.3GB)
3. Detect your Neovim and tmux configurations
4. Create initial configuration file

This only needs to be run once. Use --force to re-download the model.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&modelURL, "model-url", "", "custom model URL (default: Phi-3-mini from HuggingFace)")
	initCmd.Flags().BoolVar(&skipConfig, "skip-config", false, "skip config detection")
	initCmd.Flags().BoolVar(&forceInit, "force", false, "re-download model even if it exists")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Styles for output
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	fmt.Println(titleStyle.Render("\n🚀 Initializing Cliq...\n"))

	// Step 1: Create directories
	fmt.Println(infoStyle.Render("Creating directories..."))
	if err := createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	fmt.Println(successStyle.Render("  ✓ Directories created"))

	// Step 2: Check/download model
	cfg := config.Default()
	modelPath := cfg.GetModelPath()

	if _, err := os.Stat(modelPath); os.IsNotExist(err) || forceInit {
		fmt.Println(infoStyle.Render("\nDownloading model (this may take a while)..."))

		url := modelURL
		if url == "" {
			url = llm.DefaultModelURL
		}

		if err := llm.DownloadModel(url, modelPath); err != nil {
			return fmt.Errorf("failed to download model: %w", err)
		}
		fmt.Println(successStyle.Render("  ✓ Model downloaded successfully"))
	} else {
		fmt.Println(successStyle.Render("  ✓ Model already exists"))
	}

	// Step 3: Detect configurations
	if !skipConfig {
		fmt.Println(infoStyle.Render("\nDetecting configurations..."))

		// Detect Neovim config
		nvimPath, err := config.DetectNvimConfig()
		if err == nil {
			fmt.Printf("  ✓ Found Neovim config: %s\n", nvimPath)
			cfg.Nvim.ConfigPath = nvimPath
		} else {
			fmt.Println(warnStyle.Render("  ! Neovim config not found"))
		}

		// Detect tmux config
		tmuxPath, err := config.DetectTmuxConfig()
		if err == nil {
			fmt.Printf("  ✓ Found tmux config: %s\n", tmuxPath)
			cfg.Tmux.ConfigPath = tmuxPath
		} else {
			fmt.Println(warnStyle.Render("  ! tmux config not found"))
		}
	}

	// Step 4: Save configuration
	fmt.Println(infoStyle.Render("\nSaving configuration..."))
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	fmt.Println(successStyle.Render("  ✓ Configuration saved"))

	// Done
	fmt.Println(titleStyle.Render("\n✨ Cliq is ready to use!\n"))
	fmt.Println("Try running:")
	fmt.Println(infoStyle.Render("  cliq \"how do I delete a line in vim\""))
	fmt.Println(infoStyle.Render("  cliq \"split tmux window\""))
	fmt.Println(infoStyle.Render("  cliq -i                              # Interactive mode"))

	return nil
}

func createDirectories() error {
	// Get directory paths
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	dataDir, err := config.GetDataDir()
	if err != nil {
		return err
	}

	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return err
	}

	// Create directories
	dirs := []string{
		configDir,
		dataDir,
		filepath.Join(dataDir, "model"),
		cacheDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
