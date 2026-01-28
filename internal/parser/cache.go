package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/cliq-cli/cliq/internal/config"
)

// Cache represents cached configuration data
type Cache struct {
	NvimConfig   *NvimConfig            `json:"nvim_config,omitempty"`
	TmuxConfig   *TmuxConfig            `json:"tmux_config,omitempty"`
	LastParsed   time.Time              `json:"last_parsed"`
	ConfigHashes map[string]string      `json:"config_hashes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LoadCache loads the cache from disk
func LoadCache() (*Cache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Cache{
				ConfigHashes: make(map[string]string),
				Metadata:     make(map[string]interface{}),
			}, nil
		}
		return nil, err
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	// Initialize maps if nil
	if cache.ConfigHashes == nil {
		cache.ConfigHashes = make(map[string]string)
	}
	if cache.Metadata == nil {
		cache.Metadata = make(map[string]interface{})
	}

	return &cache, nil
}

// Save saves the cache to disk
func (c *Cache) Save() error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Update last parsed time
	c.LastParsed = time.Now()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// IsStale checks if the cache is older than the specified TTL
func (c *Cache) IsStale(ttlHours int) bool {
	if c.LastParsed.IsZero() {
		return true
	}

	ttl := time.Duration(ttlHours) * time.Hour
	return time.Since(c.LastParsed) > ttl
}

// NeedsRefresh checks if any source config files have been modified since the cache was created
func (c *Cache) NeedsRefresh() bool {
	if c.NvimConfig != nil && c.NvimConfig.ConfigPath != "" {
		if modified, _ := isFileModifiedSince(c.NvimConfig.ConfigPath, c.LastParsed); modified {
			return true
		}
	}

	if c.TmuxConfig != nil && c.TmuxConfig.ConfigPath != "" {
		if modified, _ := isFileModifiedSince(c.TmuxConfig.ConfigPath, c.LastParsed); modified {
			return true
		}
	}

	return false
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	c.NvimConfig = nil
	c.TmuxConfig = nil
	c.ConfigHashes = make(map[string]string)
	c.LastParsed = time.Time{}

	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	return os.Remove(cachePath)
}

// GetSummary returns a summary of the cached data
func (c *Cache) GetSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["last_parsed"] = c.LastParsed.Format(time.RFC3339)
	summary["is_stale_24h"] = c.IsStale(24)
	summary["needs_refresh"] = c.NeedsRefresh()

	if c.NvimConfig != nil {
		summary["nvim_keymaps_count"] = len(c.NvimConfig.Keymaps)
		summary["nvim_plugins_count"] = len(c.NvimConfig.Plugins)
		summary["nvim_leader"] = c.NvimConfig.Leader
	}

	if c.TmuxConfig != nil {
		summary["tmux_keymaps_count"] = len(c.TmuxConfig.Keymaps)
		summary["tmux_prefix"] = c.TmuxConfig.Prefix
	}

	return summary
}

// getCachePath returns the full path to the cache file
func getCachePath() (string, error) {
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "config-cache.json"), nil
}

// isFileModifiedSince checks if a file has been modified since the given time
func isFileModifiedSince(path string, since time.Time) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.ModTime().After(since), nil
}
