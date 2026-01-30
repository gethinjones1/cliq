package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/cliq-cli/cliq/internal/config"
	"github.com/cliq-cli/cliq/internal/llm"
	"github.com/cliq-cli/cliq/internal/parser"
	"github.com/cliq-cli/cliq/internal/response"
)

// executeQuery runs the query through the LLM and displays the response
func executeQuery(query string, cfg *config.Config) error {
	// Load or create cache
	var nvimConfig *parser.NvimConfig
	var tmuxConfig *parser.TmuxConfig

	noCache := viper.GetBool("no-cache")

	if !noCache && cfg.Cache.Enabled {
		cache, err := parser.LoadCache()
		if err == nil && !cache.IsStale(cfg.Cache.TTLHours) {
			nvimConfig = cache.NvimConfig
			tmuxConfig = cache.TmuxConfig
		}
	}

	// Parse configs if not cached
	if nvimConfig == nil && cfg.Nvim.ConfigPath != "" {
		var err error
		nvimConfig, err = parser.ParseNvimConfig(cfg.Nvim.ConfigPath)
		if err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: could not parse nvim config: %v\n", err)
		}
	}

	if tmuxConfig == nil && cfg.Tmux.ConfigPath != "" {
		var err error
		tmuxConfig, err = parser.ParseTmuxConfig(cfg.Tmux.ConfigPath)
		if err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: could not parse tmux config: %v\n", err)
		}
	}

	// Save to cache if enabled
	if cfg.Cache.Enabled && !noCache {
		cache := &parser.Cache{
			NvimConfig: nvimConfig,
			TmuxConfig: tmuxConfig,
		}
		if err := cache.Save(); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: could not save cache: %v\n", err)
		}
	}

	// Build prompt with configuration context
	prompt := llm.BuildPrompt(query, nvimConfig, tmuxConfig)

	// Create LLM client
	client, err := llm.NewClient(cfg.GetModelPath(), cfg.Model.OllamaModel, cfg.Model.Temperature, cfg.Model.MaxTokens)
	if err != nil {
		return fmt.Errorf("failed to initialize LLM: %w", err)
	}
	defer client.Close()

	if verbose {
		fmt.Fprintln(os.Stderr, "Query:", query)
		fmt.Fprintln(os.Stderr, "Backend:", client.GetBackend())
		if client.GetBackend() == "ollama" {
			fmt.Fprintln(os.Stderr, "Model:", cfg.Model.OllamaModel)
		}
	}

	// Generate response
	llmResponse, err := client.Query(prompt)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	// Format and display response
	format := viper.GetString("format")
	output, err := formatOutput(llmResponse, format, nvimConfig, tmuxConfig, query)
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	fmt.Println(output)
	return nil
}

// formatOutput formats the LLM response based on the specified format
func formatOutput(llmResponse, format string, nvimCfg *parser.NvimConfig, tmuxCfg *parser.TmuxConfig, query string) (string, error) {
	// Parse the LLM response
	resp := response.Parse(llmResponse)

	// Add user-specific keymaps if relevant
	if nvimCfg != nil {
		relevantKeymaps := findRelevantKeymaps(query, nvimCfg.Keymaps)
		resp.UserKeymaps = relevantKeymaps
	}

	if tmuxCfg != nil && strings.Contains(strings.ToLower(query), "tmux") {
		resp.TmuxPrefix = tmuxCfg.Prefix
	}

	switch format {
	case "json":
		return resp.ToJSON()
	case "markdown":
		return resp.ToMarkdown(), nil
	default:
		return resp.ToText(), nil
	}
}

// findRelevantKeymaps finds keymaps that might be relevant to the query
func findRelevantKeymaps(query string, keymaps []parser.Keymap) []string {
	query = strings.ToLower(query)
	var relevant []string

	keywords := extractKeywords(query)

	for _, km := range keymaps {
		desc := strings.ToLower(km.Description)
		rhs := strings.ToLower(km.Rhs)

		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) || strings.Contains(rhs, keyword) {
				relevant = append(relevant, fmt.Sprintf("%s -> %s (%s)", km.Lhs, km.Rhs, km.Description))
				break
			}
		}

		if len(relevant) >= 3 {
			break
		}
	}

	return relevant
}

// extractKeywords extracts relevant keywords from the query
func extractKeywords(query string) []string {
	// Common keywords to look for
	keywords := []string{
		"delete", "yank", "copy", "paste", "cut",
		"search", "find", "replace", "substitute",
		"split", "window", "buffer", "tab",
		"save", "write", "quit", "exit",
		"jump", "goto", "navigate", "move",
		"select", "visual", "block",
		"indent", "format", "comment",
		"fold", "unfold",
		"undo", "redo",
		"macro", "register",
		"lsp", "diagnostic", "definition", "reference",
		"telescope", "fuzzy", "file",
	}

	var found []string
	query = strings.ToLower(query)

	for _, kw := range keywords {
		if strings.Contains(query, kw) {
			found = append(found, kw)
		}
	}

	return found
}
