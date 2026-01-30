package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/cliq-cli/cliq/internal/config"
	"github.com/cliq-cli/cliq/internal/llm"
)

var (
	modelURL     string
	skipConfig   bool
	forceInit    bool
	useOllama    bool
	downloadGGUF bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Cliq - set up LLM backend and configuration",
	Long: `Initialize Cliq by setting up an LLM backend and configuration.

Cliq supports multiple LLM backends:

1. Ollama (recommended) - Easiest setup, manages models automatically
   Install: https://ollama.ai
   Then: cliq init --ollama

2. llama.cpp server - For custom GGUF models
   Run: llama-server -m your-model.gguf --port 8080
   Then: cliq init

3. Direct GGUF download - Downloads Phi-3 model (~2.3GB)
   Run: cliq init --download

This command will also detect your Neovim and tmux configurations.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&useOllama, "ollama", false, "set up with Ollama (recommended)")
	initCmd.Flags().BoolVar(&downloadGGUF, "download", false, "download GGUF model directly")
	initCmd.Flags().StringVar(&modelURL, "model-url", "", "custom model URL for --download")
	initCmd.Flags().BoolVar(&skipConfig, "skip-config", false, "skip config detection")
	initCmd.Flags().BoolVar(&forceInit, "force", false, "re-download model even if exists")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Styles for output
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	fmt.Println(titleStyle.Render("\n🚀 Initializing Cliq...\n"))

	// Step 1: Create directories
	fmt.Println(infoStyle.Render("Creating directories..."))
	if err := createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	fmt.Println(successStyle.Render("  ✓ Directories created"))

	cfg := config.Default()

	// Step 2: Set up LLM backend
	fmt.Println(infoStyle.Render("\nSetting up LLM backend..."))

	if useOllama {
		// Check if ollama is installed
		if _, err := exec.LookPath("ollama"); err != nil {
			fmt.Println(warnStyle.Render("  ! Ollama not found"))
			fmt.Println()
			fmt.Println("Install Ollama from: https://ollama.ai")
			fmt.Println("Then run: " + cmdStyle.Render("cliq init --ollama"))
			return fmt.Errorf("ollama not installed")
		}

		fmt.Println(successStyle.Render("  ✓ Ollama detected"))

		// Pull phi3 model
		fmt.Println(infoStyle.Render("  Pulling phi3 model (this may take a while)..."))
		pullCmd := exec.Command("ollama", "pull", "phi3")
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			return fmt.Errorf("failed to pull phi3 model: %w", err)
		}
		fmt.Println(successStyle.Render("  ✓ phi3 model ready"))

		cfg.Model.Backend = "ollama"

	} else if downloadGGUF {
		// Download GGUF model directly
		modelPath := cfg.GetModelPath()

		if _, err := os.Stat(modelPath); os.IsNotExist(err) || forceInit {
			fmt.Println(infoStyle.Render("  Downloading model (~2.3GB, this may take a while)..."))

			url := modelURL
			if url == "" {
				url = llm.DefaultModelURL
			}

			if err := llm.DownloadModel(url, modelPath); err != nil {
				return fmt.Errorf("failed to download model: %w", err)
			}
			fmt.Println(successStyle.Render("  ✓ Model downloaded"))
		} else {
			fmt.Println(successStyle.Render("  ✓ Model already exists"))
		}

		cfg.Model.Backend = "llama-cli"

	} else {
		// Auto-detect available backend
		backend := detectAvailableBackend()

		switch backend {
		case "ollama":
			fmt.Println(successStyle.Render("  ✓ Ollama detected and running"))
			// Check if phi3 is available
			if !checkOllamaModel("phi3") {
				fmt.Println(infoStyle.Render("  Pulling phi3 model..."))
				pullCmd := exec.Command("ollama", "pull", "phi3")
				pullCmd.Stdout = os.Stdout
				pullCmd.Stderr = os.Stderr
				if err := pullCmd.Run(); err != nil {
					fmt.Println(warnStyle.Render("  ! Failed to pull phi3, you may need to pull it manually"))
				} else {
					fmt.Println(successStyle.Render("  ✓ phi3 model ready"))
				}
			} else {
				fmt.Println(successStyle.Render("  ✓ phi3 model available"))
			}
			cfg.Model.Backend = "ollama"

		case "llama-server":
			fmt.Println(successStyle.Render("  ✓ llama-server detected and running"))
			cfg.Model.Backend = "llama-server"

		case "llama-cli":
			fmt.Println(successStyle.Render("  ✓ llama-cli detected"))
			modelPath := cfg.GetModelPath()
			if _, err := os.Stat(modelPath); os.IsNotExist(err) {
				fmt.Println(warnStyle.Render("  ! Model file not found"))
				fmt.Println()
				fmt.Println("You have llama-cli but no model. Options:")
				fmt.Println("  1. " + cmdStyle.Render("cliq init --download") + " to download Phi-3")
				fmt.Println("  2. " + cmdStyle.Render("cliq init --ollama") + " to use Ollama instead (recommended)")
				return fmt.Errorf("model not found")
			}
			cfg.Model.Backend = "llama-cli"

		default:
			fmt.Println(warnStyle.Render("  ! No LLM backend detected"))
			fmt.Println()
			fmt.Println("Please install an LLM backend:")
			fmt.Println()
			fmt.Println("Option 1 - Ollama (recommended, easiest):")
			fmt.Println("  1. Install from " + cmdStyle.Render("https://ollama.ai"))
			fmt.Println("  2. Run " + cmdStyle.Render("ollama serve") + " (or it auto-starts)")
			fmt.Println("  3. Run " + cmdStyle.Render("cliq init --ollama"))
			fmt.Println()
			fmt.Println("Option 2 - llama.cpp server:")
			fmt.Println("  1. Build llama.cpp from https://github.com/ggerganov/llama.cpp")
			fmt.Println("  2. Run " + cmdStyle.Render("llama-server -m model.gguf --port 8080"))
			fmt.Println("  3. Run " + cmdStyle.Render("cliq init"))
			fmt.Println()
			fmt.Println("Option 3 - Download model directly:")
			fmt.Println("  Run " + cmdStyle.Render("cliq init --download"))
			return fmt.Errorf("no LLM backend available")
		}
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
	fmt.Println(cmdStyle.Render("  cliq \"how do I delete a line in vim\""))
	fmt.Println(cmdStyle.Render("  cliq \"jump to the 6th f character\""))
	fmt.Println(cmdStyle.Render("  cliq \"split tmux window vertically\""))
	fmt.Println(cmdStyle.Render("  cliq -i") + "   # Interactive mode")

	return nil
}

func detectAvailableBackend() string {
	// Check for running llama-server
	if llm.CheckLlamaServerRunning() {
		return "llama-server"
	}

	// Check for ollama
	if _, err := exec.LookPath("ollama"); err == nil {
		if llm.CheckOllamaRunning() {
			return "ollama"
		}
	}

	// Check for llama-cli
	if _, err := exec.LookPath("llama-cli"); err == nil {
		return "llama-cli"
	}
	if _, err := exec.LookPath("llama"); err == nil {
		return "llama-cli"
	}

	return ""
}

func checkOllamaModel(model string) bool {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return contains(string(output), model)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func createDirectories() error {
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
