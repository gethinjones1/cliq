package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/cliq-cli/cliq/internal/config"
	"github.com/cliq-cli/cliq/internal/llm"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information including the Cliq version, model details, and system information.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	version, commit, date := GetVersionInfo()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	fmt.Println(titleStyle.Render("Cliq - AI-powered CLI assistant for Neovim and tmux"))
	fmt.Println()

	fmt.Printf("%s %s\n", labelStyle.Render("Version:"), version)
	fmt.Printf("%s %s\n", labelStyle.Render("Commit:"), commit)
	fmt.Printf("%s %s\n", labelStyle.Render("Built:"), date)
	fmt.Printf("%s %s\n", labelStyle.Render("Go:"), runtime.Version())
	fmt.Printf("%s %s/%s\n", labelStyle.Render("OS/Arch:"), runtime.GOOS, runtime.GOARCH)

	// Check model status
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	modelPath := cfg.GetModelPath()
	if info, err := os.Stat(modelPath); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("%s %s (%.1f MB)\n", labelStyle.Render("Model:"), llm.ModelName, sizeMB)
		fmt.Printf("%s %s\n", labelStyle.Render("Model Path:"), modelPath)
	} else {
		fmt.Printf("%s Not installed (run 'cliq init')\n", labelStyle.Render("Model:"))
	}
}
