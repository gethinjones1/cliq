package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "cliq"), nil
	}

	// Fall back to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "cliq"), nil
}

// GetDataDir returns the data directory path
func GetDataDir() (string, error) {
	// Check XDG_DATA_HOME first
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, "cliq"), nil
	}

	// Fall back to ~/.local/share
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "cliq"), nil
}

// GetCacheDir returns the cache directory path
func GetCacheDir() (string, error) {
	// Check XDG_CACHE_HOME first
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "cliq"), nil
	}

	// Fall back to ~/.cache
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".cache", "cliq"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() string {
	configDir, err := GetConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "config.toml")
}

// DetectNvimConfig attempts to find the Neovim configuration directory
func DetectNvimConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Paths to check in order of preference
	paths := []string{
		// XDG config
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "nvim"),
		// Standard locations
		filepath.Join(home, ".config", "nvim"),
		filepath.Join(home, ".nvim"),
	}

	// Also check NVIM_APPNAME for custom nvim configurations
	if appName := os.Getenv("NVIM_APPNAME"); appName != "" {
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(home, ".config")
		}
		paths = append([]string{filepath.Join(xdgConfig, appName)}, paths...)
	}

	for _, path := range paths {
		if path == "" {
			continue
		}

		// Check for init.lua or init.vim
		initLua := filepath.Join(path, "init.lua")
		initVim := filepath.Join(path, "init.vim")

		if _, err := os.Stat(initLua); err == nil {
			return path, nil
		}
		if _, err := os.Stat(initVim); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("neovim configuration not found")
}

// DetectTmuxConfig attempts to find the tmux configuration file
func DetectTmuxConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Paths to check in order of preference
	paths := []string{
		// XDG config
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "tmux", "tmux.conf"),
		// Standard locations
		filepath.Join(home, ".config", "tmux", "tmux.conf"),
		filepath.Join(home, ".tmux.conf"),
	}

	for _, path := range paths {
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("tmux configuration not found")
}

// DetectAllConfigs attempts to detect both nvim and tmux configurations
func DetectAllConfigs() (nvimPath, tmuxPath string) {
	nvimPath, _ = DetectNvimConfig()
	tmuxPath, _ = DetectTmuxConfig()
	return
}
