package llm

import (
	"fmt"
	"strings"

	"github.com/cliq-cli/cliq/internal/parser"
)

// SystemPrompt is the base system prompt for the LLM
const SystemPrompt = `You are Cliq, an expert assistant for Neovim and tmux.
You help users learn commands, understand motions, and navigate their tools efficiently.

Core Guidelines:
1. Be concise but complete - no unnecessary verbosity
2. Show the most common/idiomatic solution first
3. Mention useful alternatives
4. Explain what commands do in simple terms
5. When user has custom keybindings, always mention them
6. Use plain language, avoid jargon unless necessary
7. Focus on practical, actionable advice

Response Structure:
1. Command: The main command(s) to use
2. Explanation: What it does in plain language
3. Alternatives: Other ways to accomplish the same thing
4. Related: Related commands that might be useful
5. Tip: A pro tip or best practice (optional)

Always format your response with clear sections using the labels:
Command:, Explanation:, Alternatives:, Related:, Tip:`

// BuildPrompt constructs the full prompt including user configuration context
func BuildPrompt(query string, nvimCfg *parser.NvimConfig, tmuxCfg *parser.TmuxConfig) string {
	var sb strings.Builder

	sb.WriteString(SystemPrompt)
	sb.WriteString("\n\n")

	// Add configuration context if available
	if nvimCfg != nil || tmuxCfg != nil {
		sb.WriteString("User's Configuration:\n")

		if nvimCfg != nil {
			sb.WriteString(fmt.Sprintf("- Leader key: %s\n", formatLeaderKey(nvimCfg.Leader)))

			if len(nvimCfg.Plugins) > 0 {
				sb.WriteString("- Detected plugins: ")
				plugins := make([]string, 0, len(nvimCfg.Plugins))
				for _, p := range nvimCfg.Plugins {
					if p.Enabled && len(plugins) < 10 {
						plugins = append(plugins, p.Name)
					}
				}
				sb.WriteString(strings.Join(plugins, ", "))
				sb.WriteString("\n")
			}

			// Add relevant keymaps (limit to avoid token overflow)
			relevantKeymaps := findRelevantKeymapsForQuery(query, nvimCfg.Keymaps, 5)
			if len(relevantKeymaps) > 0 {
				sb.WriteString("- Custom keymaps:\n")
				for _, km := range relevantKeymaps {
					sb.WriteString(fmt.Sprintf("  [%s] %s -> %s", km.Mode, km.Lhs, km.Rhs))
					if km.Description != "" {
						sb.WriteString(fmt.Sprintf(" (%s)", km.Description))
					}
					sb.WriteString("\n")
				}
			}
		}

		if tmuxCfg != nil {
			sb.WriteString(fmt.Sprintf("- Tmux prefix: %s\n", tmuxCfg.Prefix))

			// Add relevant tmux keymaps
			if strings.Contains(strings.ToLower(query), "tmux") && len(tmuxCfg.Keymaps) > 0 {
				sb.WriteString("- Custom tmux bindings:\n")
				count := 0
				for _, km := range tmuxCfg.Keymaps {
					if count >= 5 {
						break
					}
					sb.WriteString(fmt.Sprintf("  %s -> %s\n", km.Key, km.Command))
					count++
				}
			}
		}

		sb.WriteString("\nWhen relevant, mention the user's custom keybindings in your response.\n")
	}

	sb.WriteString("\n")
	sb.WriteString("User Question: ")
	sb.WriteString(query)
	sb.WriteString("\n\nResponse:")

	return sb.String()
}

// formatLeaderKey formats the leader key for display
func formatLeaderKey(leader string) string {
	switch leader {
	case " ":
		return "<Space>"
	case "\\":
		return "\\"
	case ",":
		return ","
	case "":
		return "\\ (default)"
	default:
		return leader
	}
}

// findRelevantKeymapsForQuery finds keymaps that might be relevant to the query
func findRelevantKeymapsForQuery(query string, keymaps []parser.Keymap, limit int) []parser.Keymap {
	query = strings.ToLower(query)
	var relevant []parser.Keymap

	// Keywords to look for
	keywords := extractQueryKeywords(query)

	for _, km := range keymaps {
		if len(relevant) >= limit {
			break
		}

		desc := strings.ToLower(km.Description)
		rhs := strings.ToLower(km.Rhs)
		lhs := strings.ToLower(km.Lhs)

		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) ||
				strings.Contains(rhs, keyword) ||
				strings.Contains(lhs, keyword) {
				relevant = append(relevant, km)
				break
			}
		}
	}

	return relevant
}

// extractQueryKeywords extracts relevant keywords from the query
func extractQueryKeywords(query string) []string {
	// Map of query terms to vim/tmux keywords
	keywordMap := map[string][]string{
		"delete":    {"delete", "d", "dd", "del", "remove"},
		"yank":      {"yank", "y", "yy", "copy"},
		"copy":      {"yank", "y", "copy", "clipboard"},
		"paste":     {"paste", "p", "put"},
		"search":    {"search", "/", "find", "grep", "telescope"},
		"replace":   {"replace", "substitute", "s/", "%s/"},
		"split":     {"split", "vsplit", "sp", "vs", "window"},
		"window":    {"window", "split", "vsplit", "wincmd"},
		"buffer":    {"buffer", "buf", "bn", "bp"},
		"tab":       {"tab", "tabnew", "tabclose"},
		"save":      {"save", "write", "w", "update"},
		"quit":      {"quit", "q", "exit", "close"},
		"jump":      {"jump", "goto", "go", "navigate"},
		"fold":      {"fold", "unfold", "za", "zo", "zc"},
		"undo":      {"undo", "u", "redo"},
		"macro":     {"macro", "record", "q", "@"},
		"lsp":       {"lsp", "diagnostic", "definition", "reference", "hover"},
		"telescope": {"telescope", "find_files", "grep", "fuzzy"},
		"comment":   {"comment", "gcc", "gc"},
		"indent":    {"indent", ">>", "<<", "="},
		"visual":    {"visual", "v", "V", "select"},
		"tmux":      {"tmux", "prefix", "pane", "session"},
	}

	var keywords []string
	seen := make(map[string]bool)

	for term, kws := range keywordMap {
		if strings.Contains(query, term) {
			for _, kw := range kws {
				if !seen[kw] {
					keywords = append(keywords, kw)
					seen[kw] = true
				}
			}
		}
	}

	return keywords
}
