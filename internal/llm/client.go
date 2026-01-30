package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client wraps the LLM inference engine
type Client struct {
	modelPath   string
	temperature float64
	maxTokens   int
	backend     string // "llama-server", "ollama", "llama-cli"
	serverURL   string
}

// NewClient creates a new LLM client and auto-detects the best available backend
func NewClient(modelPath string, temperature float64, maxTokens int) (*Client, error) {
	client := &Client{
		modelPath:   modelPath,
		temperature: temperature,
		maxTokens:   maxTokens,
	}

	// Try to detect the best available backend
	backend, serverURL := detectBackend(modelPath)
	client.backend = backend
	client.serverURL = serverURL

	if backend == "" {
		return nil, fmt.Errorf("no LLM backend available. Please install one of:\n" +
			"  1. ollama (recommended): https://ollama.ai\n" +
			"  2. llama.cpp server: https://github.com/ggerganov/llama.cpp\n" +
			"  3. llama-cli from llama.cpp")
	}

	return client, nil
}

// detectBackend finds the best available LLM backend
func detectBackend(modelPath string) (backend string, serverURL string) {
	// 1. Check if llama-server is running
	if url := checkLlamaServer(); url != "" {
		return "llama-server", url
	}

	// 2. Check for ollama
	if _, err := exec.LookPath("ollama"); err == nil {
		// Check if ollama is running
		if checkOllamaRunning() {
			return "ollama", "http://localhost:11434"
		}
	}

	// 3. Check for llama-cli
	if path, err := exec.LookPath("llama-cli"); err == nil {
		if _, err := os.Stat(modelPath); err == nil {
			return "llama-cli:" + path, ""
		}
	}

	// 4. Check for llama (older name)
	if path, err := exec.LookPath("llama"); err == nil {
		if _, err := os.Stat(modelPath); err == nil {
			return "llama-cli:" + path, ""
		}
	}

	// 5. Check for local llama-server binary that's not running
	if path, err := exec.LookPath("llama-server"); err == nil {
		if _, err := os.Stat(modelPath); err == nil {
			return "llama-server-start:" + path, ""
		}
	}

	return "", ""
}

// CheckLlamaServerRunning checks if llama-server is running
func CheckLlamaServerRunning() bool {
	return checkLlamaServer() != ""
}

// CheckOllamaRunning is the exported version of checkOllamaRunning
func CheckOllamaRunning() bool {
	return checkOllamaRunning()
}

// checkLlamaServer checks if llama-server is running on common ports
func checkLlamaServer() string {
	ports := []string{"8080", "8000", "5000"}
	client := &http.Client{Timeout: 500 * time.Millisecond}

	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%s/health", port)
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return fmt.Sprintf("http://localhost:%s", port)
			}
		}
	}
	return ""
}

// checkOllamaRunning checks if ollama is running
func checkOllamaRunning() bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err == nil {
		resp.Body.Close()
		return resp.StatusCode == 200
	}
	return false
}

// Query sends a prompt to the LLM and returns the response
func (c *Client) Query(prompt string) (string, error) {
	switch {
	case c.backend == "llama-server":
		return c.queryLlamaServer(prompt)
	case c.backend == "ollama":
		return c.queryOllama(prompt)
	case strings.HasPrefix(c.backend, "llama-cli:"):
		path := strings.TrimPrefix(c.backend, "llama-cli:")
		return c.queryLlamaCLI(path, prompt)
	case strings.HasPrefix(c.backend, "llama-server-start:"):
		return "", fmt.Errorf("llama-server is installed but not running.\n" +
			"Start it with: llama-server -m %s --port 8080\n" +
			"Or use ollama instead: ollama run phi3", c.modelPath)
	default:
		return "", fmt.Errorf("no LLM backend configured")
	}
}

// queryLlamaServer queries the llama.cpp server API
func (c *Client) queryLlamaServer(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"prompt":      prompt,
		"n_predict":   c.maxTokens,
		"temperature": c.temperature,
		"stop":        []string{"\n\nUser:", "\n\nQuestion:", "```\n\n"},
		"stream":      false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.serverURL+"/completion", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("llama-server request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return strings.TrimSpace(result.Content), nil
}

// queryOllama queries the Ollama API
func (c *Client) queryOllama(prompt string) (string, error) {
	// Use phi3 model by default (can be configured)
	model := "phi3"
	if os.Getenv("CLIQ_OLLAMA_MODEL") != "" {
		model = os.Getenv("CLIQ_OLLAMA_MODEL")
	}

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": c.temperature,
			"num_predict": c.maxTokens,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(c.serverURL+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("model '%s' not found in ollama. Pull it with: ollama pull %s", model, model)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("ollama error: %s", result.Error)
	}

	return strings.TrimSpace(result.Response), nil
}

// queryLlamaCLI uses the llama.cpp CLI for inference
func (c *Client) queryLlamaCLI(llamaPath, prompt string) (string, error) {
	args := []string{
		"-m", c.modelPath,
		"-p", prompt,
		"-n", fmt.Sprintf("%d", c.maxTokens),
		"--temp", fmt.Sprintf("%.2f", c.temperature),
		"--no-display-prompt",
		"-c", "4096",
	}

	cmd := exec.Command(llamaPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("llama inference failed: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Close releases resources held by the client
func (c *Client) Close() error {
	return nil
}

// GetBackend returns the current backend being used
func (c *Client) GetBackend() string {
	return c.backend
}
