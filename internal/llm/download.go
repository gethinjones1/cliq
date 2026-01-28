package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

const (
	// ModelName is the name of the default model
	ModelName = "Phi-3-mini-4k-instruct-q4"

	// DefaultModelURL is the URL to download the model from HuggingFace
	DefaultModelURL = "https://huggingface.co/microsoft/Phi-3-mini-4k-instruct-gguf/resolve/main/Phi-3-mini-4k-instruct-q4.gguf"

	// ModelSize is the approximate size of the model in bytes (2.3GB)
	ModelSize = 2_300_000_000

	// ExpectedSHA256 is the expected checksum of the model file
	// This should be updated when the model version changes
	ExpectedSHA256 = ""
)

// DownloadModel downloads the model from the given URL to the specified path
func DownloadModel(url, destPath string) error {
	// Create the destination directory if it doesn't exist
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a temporary file for downloading
	tmpPath := destPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath) // Clean up temp file on error
	}()

	// Create HTTP client with timeout
	client := &http.Client{}

	// Make the request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Get file size for progress bar
	size := resp.ContentLength
	if size <= 0 {
		size = ModelSize
	}

	// Create progress bar
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stderr)
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Download with progress
	_, err = io.Copy(io.MultiWriter(tmpFile, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("download interrupted: %w", err)
	}

	// Close the temp file before renaming
	tmpFile.Close()

	// Verify checksum if we have one
	if ExpectedSHA256 != "" {
		checksum, err := calculateSHA256(tmpPath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
		if checksum != ExpectedSHA256 {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", ExpectedSHA256, checksum)
		}
	}

	// Rename temp file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to finalize download: %w", err)
	}

	return nil
}

// VerifyModel verifies the model file exists and has the correct checksum
func VerifyModel(path string) error {
	// Check file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("model file not found")
		}
		return err
	}

	// Check file size is reasonable
	if info.Size() < 100_000_000 { // Less than 100MB is suspicious
		return fmt.Errorf("model file appears to be corrupted (too small)")
	}

	// Verify checksum if we have one
	if ExpectedSHA256 != "" {
		checksum, err := calculateSHA256(path)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
		if checksum != ExpectedSHA256 {
			return fmt.Errorf("checksum mismatch: model may be corrupted")
		}
	}

	return nil
}

// calculateSHA256 calculates the SHA256 checksum of a file
func calculateSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// GetModelInfo returns information about the model
func GetModelInfo(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":     ModelName,
		"path":     path,
		"size":     info.Size(),
		"size_mb":  float64(info.Size()) / (1024 * 1024),
		"modified": info.ModTime(),
	}, nil
}
