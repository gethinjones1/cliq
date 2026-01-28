package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration
type Config struct {
	General GeneralConfig `toml:"general"`
	Model   ModelConfig   `toml:"model"`
	Nvim    NvimConfig    `toml:"nvim"`
	Tmux    TmuxConfig    `toml:"tmux"`
	Cache   CacheConfig   `toml:"cache"`
	TUI     TUIConfig     `toml:"tui"`
}

// GeneralConfig holds general application settings
type GeneralConfig struct {
	ResponseStyle string `toml:"response_style"` // concise, detailed, minimal
}

// ModelConfig holds model-related settings
type ModelConfig struct {
	Path        string  `toml:"path"`
	AutoUpdate  bool    `toml:"auto_update"`
	Temperature float64 `toml:"temperature"`
	MaxTokens   int     `toml:"max_tokens"`
}

// NvimConfig holds Neovim-related settings
type NvimConfig struct {
	ConfigPath     string   `toml:"config_path"`
	AutoDetect     bool     `toml:"auto_detect"`
	ParsePlugins   bool     `toml:"parse_plugins"`
	TrackedPlugins []string `toml:"tracked_plugins"`
}

// TmuxConfig holds tmux-related settings
type TmuxConfig struct {
	ConfigPath string `toml:"config_path"`
	AutoDetect bool   `toml:"auto_detect"`
}

// CacheConfig holds caching settings
type CacheConfig struct {
	Enabled  bool   `toml:"enabled"`
	TTLHours int    `toml:"ttl_hours"`
	Path     string `toml:"path"`
}

// TUIConfig holds TUI-related settings
type TUIConfig struct {
	Mouse    bool   `toml:"mouse"`
	Theme    string `toml:"theme"` // auto, light, dark
	ShowTips bool   `toml:"show_tips"`
}

// Default returns a configuration with default values
func Default() *Config {
	dataDir, _ := GetDataDir()
	cacheDir, _ := GetCacheDir()

	return &Config{
		General: GeneralConfig{
			ResponseStyle: "concise",
		},
		Model: ModelConfig{
			Path:        filepath.Join(dataDir, "model", "phi-3-mini-q4.gguf"),
			AutoUpdate:  false,
			Temperature: 0.7,
			MaxTokens:   512,
		},
		Nvim: NvimConfig{
			ConfigPath:   "",
			AutoDetect:   true,
			ParsePlugins: true,
			TrackedPlugins: []string{
				"telescope.nvim",
				"nvim-lspconfig",
				"nvim-tree.lua",
				"which-key.nvim",
			},
		},
		Tmux: TmuxConfig{
			ConfigPath: "",
			AutoDetect: true,
		},
		Cache: CacheConfig{
			Enabled:  true,
			TTLHours: 24,
			Path:     cacheDir,
		},
		TUI: TUIConfig{
			Mouse:    true,
			Theme:    "auto",
			ShowTips: true,
		},
	}
}

// Load loads the configuration from file
func Load() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetModelPath returns the full path to the model file
func (c *Config) GetModelPath() string {
	if c.Model.Path != "" {
		return expandPath(c.Model.Path)
	}

	dataDir, _ := GetDataDir()
	return filepath.Join(dataDir, "model", "phi-3-mini-q4.gguf")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
