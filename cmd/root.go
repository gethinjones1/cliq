package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cliq-cli/cliq/internal/config"
)

var (
	cfgFile     string
	verbose     bool
	versionInfo struct {
		Version string
		Commit  string
		Date    string
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cliq [query]",
	Short: "AI-powered CLI assistant for Neovim and tmux",
	Long: `Cliq is a privacy-first CLI tool that provides AI-powered assistance
for Neovim and tmux commands. It runs entirely locally using a small
language model and is aware of your personal configuration.

Examples:
  cliq "how do I delete a line"
  cliq "split tmux window vertically"
  cliq "search and replace in visual mode"
  cliq -i                              # Interactive mode`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRootCmd,
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	// Check if interactive mode
	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		return runInteractive()
	}

	if len(args) == 0 {
		return cmd.Help()
	}
	return runQuery(args[0])
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets the version information from main
func SetVersionInfo(version, commit, date string) {
	versionInfo.Version = version
	versionInfo.Commit = commit
	versionInfo.Date = date
}

// GetVersionInfo returns the version information
func GetVersionInfo() (version, commit, date string) {
	return versionInfo.Version, versionInfo.Commit, versionInfo.Date
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/cliq/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Query-specific flags
	rootCmd.Flags().StringP("format", "f", "text", "output format (text|json|markdown)")
	rootCmd.Flags().Bool("no-cache", false, "skip config cache")
	rootCmd.Flags().BoolP("interactive", "i", false, "launch interactive TUI mode")

	// Bind flags to viper
	viper.BindPFlag("format", rootCmd.Flags().Lookup("format"))
	viper.BindPFlag("no-cache", rootCmd.Flags().Lookup("no-cache"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory
		configDir, err := config.GetConfigDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not determine config directory:", err)
			return
		}

		// Search config in config directory with name "config" (without extension).
		viper.AddConfigPath(configDir)
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

// runQuery handles the main query execution
func runQuery(query string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		}
		cfg = config.Default()
	}

	// Check if model exists
	modelPath := cfg.GetModelPath()
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Println("Model not found. Please run 'cliq init' first to download the model.")
		return fmt.Errorf("model not found at %s", modelPath)
	}

	// Execute query using LLM
	return executeQuery(query, cfg)
}
