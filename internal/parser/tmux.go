package parser

import (
	"os"
	"regexp"
	"strings"
)

// TmuxConfig represents parsed tmux configuration
type TmuxConfig struct {
	Prefix     string
	Keymaps    []TmuxKeymap
	ConfigPath string
	Options    map[string]string
}

// TmuxKeymap represents a tmux key binding
type TmuxKeymap struct {
	Key         string
	Command     string
	Description string
	Table       string // key table (prefix, root, copy-mode, etc.)
}

// ParseTmuxConfig parses a tmux configuration file
func ParseTmuxConfig(configPath string) (*TmuxConfig, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &TmuxConfig{
		ConfigPath: configPath,
		Prefix:     "C-b", // Default tmux prefix
		Keymaps:    []TmuxKeymap{},
		Options:    make(map[string]string),
	}

	text := string(content)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle line continuations
		for strings.HasSuffix(line, "\\") {
			line = strings.TrimSuffix(line, "\\")
		}

		// Parse the line
		cfg.parseLine(line)
	}

	return cfg, nil
}

// parseLine parses a single line of tmux configuration
func (cfg *TmuxConfig) parseLine(line string) {
	// Extract prefix key
	if strings.Contains(line, "prefix") {
		cfg.extractPrefix(line)
	}

	// Extract key bindings
	if strings.HasPrefix(line, "bind-key") || strings.HasPrefix(line, "bind") {
		cfg.extractBinding(line)
	}

	// Extract set-option for various settings
	if strings.HasPrefix(line, "set-option") || strings.HasPrefix(line, "set ") ||
		strings.HasPrefix(line, "set-window-option") || strings.HasPrefix(line, "setw ") {
		cfg.extractOption(line)
	}
}

// extractPrefix extracts the prefix key setting
func (cfg *TmuxConfig) extractPrefix(line string) {
	// Pattern: set -g prefix C-a
	// or: set-option -g prefix C-a
	patterns := []string{
		`(?:set-option|set)\s+(?:-g\s+)?prefix\s+(\S+)`,
		`(?:set-option|set)\s+(?:-g\s+)?prefix2\s+(\S+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			cfg.Prefix = matches[1]
			return
		}
	}
}

// extractBinding extracts a key binding
func (cfg *TmuxConfig) extractBinding(line string) {
	// Patterns for bind-key:
	// bind-key [-cnr] [-t mode-table] [-T key-table] key command [arguments]
	// bind key command [arguments]

	// Remove bind-key or bind prefix
	line = regexp.MustCompile(`^bind(?:-key)?\s+`).ReplaceAllString(line, "")

	km := TmuxKeymap{
		Table: "prefix", // default table
	}

	// Check for key table specification
	tablePattern := `-T\s+(\S+)`
	if re := regexp.MustCompile(tablePattern); re.MatchString(line) {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			km.Table = matches[1]
		}
		line = re.ReplaceAllString(line, "")
	}

	// Check for -n flag (root table, no prefix needed)
	if strings.Contains(line, " -n ") || strings.HasPrefix(line, "-n ") {
		km.Table = "root"
		line = strings.Replace(line, " -n ", " ", 1)
		line = strings.TrimPrefix(line, "-n ")
	}

	// Check for -r flag (repeatable)
	line = strings.Replace(line, " -r ", " ", -1)
	line = strings.TrimPrefix(line, "-r ")

	// Remove other flags
	line = regexp.MustCompile(`-[cnt]\s+`).ReplaceAllString(line, "")

	// Now parse the remaining: key command [args]
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	if len(parts) < 2 {
		return
	}

	km.Key = parts[0]
	km.Command = strings.TrimSpace(parts[1])

	// Try to generate a description from the command
	km.Description = describeCommand(km.Command)

	cfg.Keymaps = append(cfg.Keymaps, km)
}

// extractOption extracts a set-option or set command
func (cfg *TmuxConfig) extractOption(line string) {
	// Remove the command prefix
	line = regexp.MustCompile(`^(?:set-option|set-window-option|set|setw)\s+`).ReplaceAllString(line, "")

	// Remove flags
	line = regexp.MustCompile(`-[gswu]+\s+`).ReplaceAllString(line, "")

	// Parse option=value or option value
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	if len(parts) >= 2 {
		cfg.Options[parts[0]] = strings.Trim(parts[1], "\"'")
	} else if len(parts) == 1 && strings.Contains(parts[0], "=") {
		kv := strings.SplitN(parts[0], "=", 2)
		if len(kv) == 2 {
			cfg.Options[kv[0]] = strings.Trim(kv[1], "\"'")
		}
	}
}

// describeCommand generates a human-readable description for a tmux command
func describeCommand(cmd string) string {
	cmd = strings.ToLower(cmd)

	descriptions := map[string]string{
		"split-window -h":       "Split pane horizontally",
		"split-window -v":       "Split pane vertically",
		"split-window":          "Split pane",
		"new-window":            "Create new window",
		"kill-pane":             "Close current pane",
		"kill-window":           "Close current window",
		"select-pane -L":        "Select pane to the left",
		"select-pane -R":        "Select pane to the right",
		"select-pane -U":        "Select pane above",
		"select-pane -D":        "Select pane below",
		"resize-pane":           "Resize pane",
		"next-window":           "Go to next window",
		"previous-window":       "Go to previous window",
		"last-window":           "Go to last window",
		"copy-mode":             "Enter copy mode",
		"paste-buffer":          "Paste from buffer",
		"source-file":           "Reload config",
		"command-prompt":        "Open command prompt",
		"display-message":       "Display message",
		"clock-mode":            "Show clock",
		"choose-tree":           "Choose session/window",
		"choose-session":        "Choose session",
		"choose-window":         "Choose window",
		"detach-client":         "Detach from session",
		"rename-window":         "Rename current window",
		"rename-session":        "Rename current session",
		"swap-pane":             "Swap panes",
		"rotate-window":         "Rotate panes",
		"break-pane":            "Break pane into window",
		"join-pane":             "Join pane to window",
		"send-keys":             "Send keys to pane",
		"send-prefix":           "Send prefix key",
		"set-option":            "Set option",
		"show-options":          "Show options",
		"list-keys":             "List key bindings",
		"list-sessions":         "List sessions",
		"list-windows":          "List windows",
		"list-panes":            "List panes",
	}

	// Check for exact matches first
	for pattern, desc := range descriptions {
		if strings.HasPrefix(cmd, pattern) {
			return desc
		}
	}

	// Check for partial matches
	if strings.Contains(cmd, "split") {
		return "Split pane"
	}
	if strings.Contains(cmd, "select-pane") {
		return "Select pane"
	}
	if strings.Contains(cmd, "resize") {
		return "Resize pane"
	}
	if strings.Contains(cmd, "window") {
		return "Window operation"
	}
	if strings.Contains(cmd, "pane") {
		return "Pane operation"
	}

	return ""
}
