package llm

import (
	"fmt"
	"strings"

	"github.com/cliq-cli/cliq/internal/parser"
)

// SystemPrompt is the base system prompt for the LLM
const SystemPrompt = `You are Cliq, an expert assistant for Neovim and tmux.

CRITICAL RULES:
1. Only suggest commands you are CERTAIN exist. Never invent commands.
2. Keep explanations SHORT - 1-2 sentences max.
3. The Command section must contain ONLY the exact keys to press, nothing else.
4. Do not speculate about plugins or configurations unless asked.

=== VIM/NEOVIM FUNDAMENTALS ===
Counts: Most motions accept a count prefix. Examples:
- 5j = move down 5 lines, 10k = move up 10 lines
- 3w = move forward 3 words, 2b = move back 2 words
- 4dd = delete 4 lines, 3yy = yank 3 lines

Motions:
- h/j/k/l = left/down/up/right
- w/W = next word (W includes punctuation)
- b/B = previous word
- e/E = end of word
- 0/^ = start of line / first non-blank
- $ = end of line
- gg/G = start/end of file
- {/} = paragraph up/down
- %  = matching bracket
- f{char}/F{char} = find char forward/backward on line
- t{char}/T{char} = till char forward/backward
- / = search forward, ? = search backward
- n/N = next/previous search result
- * = search word under cursor

Operators (combine with motions):
- d = delete (d + motion, dd = line, D = to end of line)
- y = yank/copy (y + motion, yy = line)
- c = change (delete + insert mode)
- > / < = indent/dedent
- = = auto-indent
- gU/gu = uppercase/lowercase

Common commands:
- :w = save, :q = quit, :wq = save and quit
- :e {file} = edit file
- :%s/old/new/g = replace all in file
- :s/old/new/g = replace all in line
- u = undo, Ctrl-r = redo
- . = repeat last change
- p/P = paste after/before
- o/O = new line below/above
- A/I = append end/insert start of line
- v/V/Ctrl-v = visual/line/block mode
- zz = center screen on cursor

=== TMUX FUNDAMENTALS ===
Default prefix: Ctrl-b (shown as C-b or prefix)

After prefix:
- c = new window
- n/p = next/previous window
- 0-9 = select window by number
- % = vertical split
- " = horizontal split
- arrow keys = move between panes
- z = toggle pane zoom
- d = detach
- [ = copy mode (then use vim keys to navigate)
- : = command mode

=== RESPONSE FORMAT ===
Command: [the exact command]
Explanation: [what it does, 1-2 sentences]
Alternatives: [other ways, if any]
Related: [related useful commands]
Tip: [optional pro tip]

=== EXAMPLES ===

Q: how do I delete 3 lines
Command: 3dd
Explanation: Deletes 3 lines starting from the cursor. The deleted text is saved to the default register.
Alternatives: d2j (delete current + 2 below), V2jd (visual select then delete)
Related: yy (yank line), p (paste), u (undo)

Q: how do I move up 50 lines
Command: 50k
Explanation: Moves the cursor up 50 lines. The number prefix works with any motion.
Alternatives: 50<Up> (arrow key also works with count)
Related: 50j (down 50 lines), gg (top of file), G (bottom of file)

Q: how to go to line 100
Command: 100G
Explanation: Jumps directly to line 100. G goes to a line number when prefixed with a count.
Alternatives: :100<Enter> (command mode)
Related: gg (line 1), G (last line), Ctrl-g (show current line number)

Q: how do I split tmux pane vertically
Command: prefix + %
Explanation: Splits the current pane vertically (side by side). Default prefix is Ctrl-b.
Alternatives: tmux split-window -h (from command line)
Related: prefix + " (horizontal split), prefix + arrow (move between panes)

Q: copy 5 lines in vim
Command: 5yy
Explanation: Yanks (copies) 5 lines starting from the cursor into the default register.
Alternatives: V4jy (visual select 5 lines then yank)
Related: p (paste below), P (paste above), "+y (yank to system clipboard)`

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
		"delete":     {"delete", "d", "dd", "del", "remove"},
		"yank":       {"yank", "y", "yy", "copy"},
		"copy":       {"yank", "y", "copy", "clipboard"},
		"paste":      {"paste", "p", "put"},
		"search":     {"search", "/", "find", "grep", "telescope"},
		"replace":    {"replace", "substitute", "s/", "%s/"},
		"split":      {"split", "vsplit", "sp", "vs", "window"},
		"window":     {"window", "split", "vsplit", "wincmd"},
		"buffer":     {"buffer", "buf", "bn", "bp"},
		"tab":        {"tab", "tabnew", "tabclose"},
		"save":       {"save", "write", "w", "update"},
		"quit":       {"quit", "q", "exit", "close"},
		"jump":       {"jump", "goto", "go", "navigate"},
		"fold":       {"fold", "unfold", "za", "zo", "zc"},
		"undo":       {"undo", "u", "redo"},
		"macro":      {"macro", "record", "q", "@"},
		"lsp":        {"lsp", "diagnostic", "definition", "reference", "hover"},
		"telescope":  {"telescope", "find_files", "grep", "fuzzy"},
		"comment":    {"comment", "gcc", "gc"},
		"indent":     {"indent", ">>", "<<", "="},
		"visual":     {"visual", "v", "V", "select"},
		"tmux":       {"tmux", "prefix", "pane", "session"},
		"debug":      {"debug", "dap", "breakpoint", "step", "continue", "terminate"},
		"breakpoint": {"breakpoint", "dap", "debug"},
		"test":       {"test", "debug", "dap"},
		"navigate":   {"navigate", "tmux", "pane", "window", "split"},
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
