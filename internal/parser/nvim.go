package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// NvimConfig represents parsed Neovim configuration
type NvimConfig struct {
	Leader     string
	Keymaps    []Keymap
	Plugins    []Plugin
	ConfigPath string
}

// Keymap represents a Neovim keymap
type Keymap struct {
	Mode        string // "n", "v", "i", etc.
	Lhs         string // Key combination
	Rhs         string // Command
	Description string
	Source      string // File where defined
}

// Plugin represents a Neovim plugin
type Plugin struct {
	Name    string
	Enabled bool
	Config  map[string]interface{}
}

// ParseNvimConfig parses the Neovim configuration directory
func ParseNvimConfig(configPath string) (*NvimConfig, error) {
	cfg := &NvimConfig{
		ConfigPath: configPath,
		Leader:     "\\", // Default leader
		Keymaps:    []Keymap{},
		Plugins:    []Plugin{},
	}

	// Check for init.lua
	initLua := filepath.Join(configPath, "init.lua")
	if _, err := os.Stat(initLua); err == nil {
		if err := cfg.parseLuaConfig(initLua); err != nil {
			// Continue even if parsing fails, we might get partial results
		}
	}

	// Check for init.vim
	initVim := filepath.Join(configPath, "init.vim")
	if _, err := os.Stat(initVim); err == nil {
		if err := cfg.parseVimConfig(initVim); err != nil {
			// Continue even if parsing fails
		}
	}

	// Check for lazy.nvim plugin specs
	lazyDir := filepath.Join(configPath, "lua", "plugins")
	if _, err := os.Stat(lazyDir); err == nil {
		cfg.parseLazyPlugins(lazyDir)
	}

	// Also check common plugin directories
	alternativePluginDirs := []string{
		filepath.Join(configPath, "lua", "config", "plugins"),
		filepath.Join(configPath, "lua", "user", "plugins"),
		filepath.Join(configPath, "after", "plugin"),
	}

	for _, dir := range alternativePluginDirs {
		if _, err := os.Stat(dir); err == nil {
			cfg.parseLazyPlugins(dir)
		}
	}

	return cfg, nil
}

// parseLuaConfig parses a Lua configuration file
func (cfg *NvimConfig) parseLuaConfig(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	text := string(content)

	// Extract leader key
	cfg.extractLeaderFromLua(text)

	// Extract keymaps using regex (safer than executing Lua)
	cfg.extractKeymapsFromLua(text, filePath)

	// Try to parse with gopher-lua for more complex extractions
	cfg.parseLuaWithInterpreter(text)

	return nil
}

// extractLeaderFromLua extracts the leader key setting from Lua code
func (cfg *NvimConfig) extractLeaderFromLua(content string) {
	// Pattern: vim.g.mapleader = "..."
	patterns := []string{
		`vim\.g\.mapleader\s*=\s*["'](.+?)["']`,
		`vim\.g\["mapleader"\]\s*=\s*["'](.+?)["']`,
		`g\.mapleader\s*=\s*["'](.+?)["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			cfg.Leader = matches[1]
			return
		}
	}
}

// extractKeymapsFromLua extracts keymap definitions from Lua code
func (cfg *NvimConfig) extractKeymapsFromLua(content, source string) {
	// Pattern: vim.keymap.set("mode", "lhs", "rhs" or function, opts)
	pattern := `vim\.keymap\.set\s*\(\s*["']([nvixsotc]+)["']\s*,\s*["']([^"']+)["']\s*,\s*(?:["']([^"']+)["']|function|[^,]+)\s*(?:,\s*\{([^}]*)\})?`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		km := Keymap{
			Mode:   match[1],
			Lhs:    match[2],
			Rhs:    match[3],
			Source: source,
		}

		// Try to extract description from opts
		if len(match) > 4 && match[4] != "" {
			descPattern := `desc\s*=\s*["']([^"']+)["']`
			descRe := regexp.MustCompile(descPattern)
			if descMatch := descRe.FindStringSubmatch(match[4]); len(descMatch) > 1 {
				km.Description = descMatch[1]
			}
		}

		cfg.Keymaps = append(cfg.Keymaps, km)
	}

	// Also look for map/noremap style (using vim.cmd)
	cmdPattern := `vim\.cmd\s*\[\[?\s*([nvixsotc]?n?o?remap)\s+([^\s]+)\s+([^\]]+)`
	cmdRe := regexp.MustCompile(cmdPattern)
	cmdMatches := cmdRe.FindAllStringSubmatch(content, -1)

	for _, match := range cmdMatches {
		if len(match) < 4 {
			continue
		}

		mode := "n"
		cmd := match[1]
		if len(cmd) > 0 && strings.Contains("nvixsotc", string(cmd[0])) {
			mode = string(cmd[0])
		}

		km := Keymap{
			Mode:   mode,
			Lhs:    strings.TrimSpace(match[2]),
			Rhs:    strings.TrimSpace(match[3]),
			Source: source,
		}

		cfg.Keymaps = append(cfg.Keymaps, km)
	}
}

// parseLuaWithInterpreter uses gopher-lua for safer evaluation
func (cfg *NvimConfig) parseLuaWithInterpreter(content string) {
	L := lua.NewState()
	defer L.Close()

	// Create a mock vim global
	vim := L.NewTable()
	g := L.NewTable()
	vim.RawSetString("g", g)
	vim.RawSetString("keymap", L.NewTable())
	vim.RawSetString("api", L.NewTable())
	vim.RawSetString("fn", L.NewTable())
	vim.RawSetString("opt", L.NewTable())
	vim.RawSetString("o", L.NewTable())
	vim.RawSetString("cmd", L.NewFunction(func(L *lua.LState) int { return 0 }))

	L.SetGlobal("vim", vim)

	// We don't actually execute the Lua code since it may have Neovim-specific
	// functions. The regex extraction above handles most cases.
}

// parseVimConfig parses a Vimscript configuration file
func (cfg *NvimConfig) parseVimConfig(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	text := string(content)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(line, "\"") {
			continue
		}

		// Extract leader key
		if strings.Contains(line, "mapleader") {
			pattern := `let\s+(?:g:)?mapleader\s*=\s*["'](.+?)["']`
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				cfg.Leader = matches[1]
			}
		}

		// Extract keymaps
		mapPattern := `^([nvixsotc]?)(?:nore)?map\s+(?:<[^>]+>\s+)?(\S+)\s+(.+)$`
		mapRe := regexp.MustCompile(mapPattern)
		if matches := mapRe.FindStringSubmatch(line); len(matches) > 3 {
			mode := matches[1]
			if mode == "" {
				mode = "n"
			}

			km := Keymap{
				Mode:   mode,
				Lhs:    matches[2],
				Rhs:    strings.TrimSpace(matches[3]),
				Source: filePath,
			}

			cfg.Keymaps = append(cfg.Keymaps, km)
		}
	}

	return nil
}

// parseLazyPlugins parses lazy.nvim plugin specifications
func (cfg *NvimConfig) parseLazyPlugins(pluginDir string) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}

		filePath := filepath.Join(pluginDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		text := string(content)

		// Extract plugin names from lazy.nvim format
		// Pattern: "username/repo-name" or 'username/repo-name'
		pluginPattern := `["']([a-zA-Z0-9_-]+/[a-zA-Z0-9._-]+)["']`
		re := regexp.MustCompile(pluginPattern)

		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				name := match[1]
				// Extract just the repo name
				parts := strings.Split(name, "/")
				if len(parts) == 2 {
					plugin := Plugin{
						Name:    parts[1],
						Enabled: !strings.Contains(text, "enabled = false"),
					}
					cfg.Plugins = append(cfg.Plugins, plugin)
				}
			}
		}

		// Also extract keymaps from plugin configs
		cfg.extractKeymapsFromLua(text, filePath)
	}
}
