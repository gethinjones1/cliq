package llm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client wraps the LLM inference engine
type Client struct {
	modelPath   string
	temperature float64
	maxTokens   int
}

// NewClient creates a new LLM client
func NewClient(modelPath string, temperature float64, maxTokens int) (*Client, error) {
	// Verify model file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model file not found: %s", modelPath)
	}

	return &Client{
		modelPath:   modelPath,
		temperature: temperature,
		maxTokens:   maxTokens,
	}, nil
}

// Query sends a prompt to the LLM and returns the response
func (c *Client) Query(prompt string) (string, error) {
	// Try to use llama-cli if available
	if llamaPath, err := exec.LookPath("llama-cli"); err == nil {
		return c.queryWithLlamaCLI(llamaPath, prompt)
	}

	// Try llama.cpp main binary
	if llamaPath, err := exec.LookPath("llama"); err == nil {
		return c.queryWithLlamaCLI(llamaPath, prompt)
	}

	// Fall back to mock response for development/testing
	return c.mockQuery(prompt)
}

// queryWithLlamaCLI uses the llama.cpp CLI for inference
func (c *Client) queryWithLlamaCLI(llamaPath, prompt string) (string, error) {
	args := []string{
		"-m", c.modelPath,
		"-p", prompt,
		"-n", fmt.Sprintf("%d", c.maxTokens),
		"--temp", fmt.Sprintf("%.2f", c.temperature),
		"--no-display-prompt",
	}

	cmd := exec.Command(llamaPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("llama inference failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// mockQuery provides responses for development/testing when no LLM is available
func (c *Client) mockQuery(prompt string) (string, error) {
	// Extract the query from the prompt
	query := strings.ToLower(prompt)

	// Provide intelligent mock responses based on common queries
	response := generateMockResponse(query)

	return response, nil
}

// Close releases resources held by the client
func (c *Client) Close() error {
	// Cleanup if needed
	return nil
}

// generateMockResponse creates helpful mock responses for development
func generateMockResponse(query string) string {
	// Check for common patterns and return appropriate responses
	switch {
	case strings.Contains(query, "delete") && strings.Contains(query, "line"):
		return `Command: dd

Explanation: Delete the current line in normal mode. The deleted content is stored in the default register, so you can paste it with 'p'.

Alternatives:
- D: Delete from cursor to end of line
- d$: Same as D
- dj: Delete current line and line below
- dk: Delete current line and line above

Related commands:
- yy: Yank (copy) current line
- cc: Change (delete and enter insert mode) current line
- S: Same as cc

Tip: Use a count prefix like '3dd' to delete 3 lines at once.`

	case strings.Contains(query, "split") && strings.Contains(query, "window"):
		return `Command: :split or :vsplit

Explanation: Split the current window horizontally (:split) or vertically (:vsplit).

Main commands:
- :split (or :sp) - Horizontal split
- :vsplit (or :vs) - Vertical split
- Ctrl-w s - Horizontal split (normal mode)
- Ctrl-w v - Vertical split (normal mode)

Navigation between splits:
- Ctrl-w h/j/k/l - Move to left/down/up/right window
- Ctrl-w w - Cycle through windows

Related commands:
- :close - Close current window
- Ctrl-w c - Close current window
- Ctrl-w o - Close all windows except current

Tip: Use :vsp filename to open a file in a new vertical split.`

	case strings.Contains(query, "search") && strings.Contains(query, "replace"):
		return `Command: :%s/old/new/g

Explanation: Search and replace 'old' with 'new' throughout the entire file.

Pattern breakdown:
- % - Entire file (can use line range like 1,10)
- s - Substitute command
- /old/ - Pattern to find
- /new/ - Replacement text
- g - Global flag (replace all occurrences on each line)

Common flags:
- g - Global (all matches on line)
- c - Confirm each replacement
- i - Case insensitive
- I - Case sensitive

Examples:
- :s/old/new/ - Replace first occurrence on current line
- :%s/old/new/gc - Replace all with confirmation
- :5,10s/old/new/g - Replace only in lines 5-10

Tip: Use \< and \> for word boundaries: :%s/\<word\>/replacement/g`

	case strings.Contains(query, "tmux") && strings.Contains(query, "split"):
		return `Command: prefix + % (vertical) or prefix + " (horizontal)

Explanation: Split the current tmux pane. Default prefix is Ctrl-b.

Split commands:
- prefix + % - Split vertically (side by side)
- prefix + " - Split horizontally (top/bottom)

Navigation:
- prefix + arrow keys - Move between panes
- prefix + o - Cycle through panes
- prefix + ; - Toggle to last active pane

Resize panes:
- prefix + Ctrl-arrow - Resize in direction
- prefix + z - Toggle pane zoom

Related commands:
- prefix + x - Close current pane
- prefix + { - Move pane left
- prefix + } - Move pane right

Tip: Use 'tmux split-window -h' or '-v' for horizontal/vertical from command line.`

	case strings.Contains(query, "copy") || strings.Contains(query, "yank"):
		return `Command: y{motion} or yy

Explanation: Yank (copy) text in Vim. Use y followed by a motion, or yy to yank the current line.

Common yank commands:
- yy - Yank current line
- y$ - Yank to end of line
- yw - Yank word
- y3w - Yank 3 words
- yip - Yank inner paragraph
- yi" - Yank inside quotes

Pasting:
- p - Paste after cursor
- P - Paste before cursor
- "0p - Paste from yank register (not affected by deletes)

Related:
- "+y - Yank to system clipboard
- "+p - Paste from system clipboard

Tip: Vim has multiple registers. Use "ay to yank into register 'a', and "ap to paste from it.`

	case strings.Contains(query, "save") || strings.Contains(query, "write"):
		return `Command: :w

Explanation: Write (save) the current buffer to disk.

Save commands:
- :w - Save current file
- :w filename - Save as new file
- :wa - Save all open buffers
- :wq - Save and quit
- :x - Save and quit (only writes if changes exist)
- ZZ - Same as :x (normal mode)

Force save:
- :w! - Force write (override read-only)
- :w !sudo tee % - Save with sudo (Linux/macOS)

Related:
- :q - Quit
- :q! - Quit without saving
- :wqa - Save all and quit

Tip: Use :set autowrite to automatically save before certain commands.`

	case strings.Contains(query, "undo") || strings.Contains(query, "redo"):
		return `Command: u (undo) or Ctrl-r (redo)

Explanation: Undo the last change with 'u', redo with Ctrl-r.

Undo commands:
- u - Undo last change
- U - Undo all changes on current line
- Ctrl-r - Redo (undo the undo)

Undo tree navigation:
- :earlier 5m - Go back 5 minutes
- :later 5m - Go forward 5 minutes
- g- - Go to older text state
- g+ - Go to newer text state

Useful settings:
- :set undofile - Persistent undo across sessions
- :set undolevels=1000 - Number of undo levels

Tip: Consider using a plugin like 'undotree' to visualize the undo history.`

	default:
		return `I can help with Neovim and tmux commands. Here are some common topics:

Navigation:
- h/j/k/l - Move cursor
- w/b - Move by word
- 0/$ - Move to line start/end
- gg/G - Move to file start/end

Editing:
- i/a - Insert mode before/after cursor
- dd - Delete line
- yy - Yank (copy) line
- p - Paste

Tmux basics:
- prefix + c - New window
- prefix + % - Vertical split
- prefix + " - Horizontal split
- prefix + d - Detach session

Try asking about specific commands like:
- "How do I delete a line?"
- "How do I split a tmux window?"
- "How do I search and replace?"`
	}
}
