package clients

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaClient struct {
	BaseURL string // e.g., "http://localhost:11434"
	Model   string // e.g., "mistral", "neural-chat", "dolphin-mixtral"
}

// NewOllamaClient creates a new Ollama client with default settings
// BaseURL defaults to http://localhost:11434
// Model defaults to mistral (a good open-source model)
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "mistral"
	}
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
	}
}

// SendPrompt sends a prompt to Ollama without a specific working directory
func (o *OllamaClient) SendPrompt(prompt string, writer io.Writer) (string, error) {
	return o.SendPromptWithDir(prompt, writer, "")
}

// SendPromptWithDir sends a prompt to Ollama (working directory is ignored for Ollama)
// Ollama doesn't support working directory context like the gemini CLI does,
// but we include it in the interface for compatibility
func (o *OllamaClient) SendPromptWithDir(prompt string, writer io.Writer, workDir string) (string, error) {
	if workDir != "" {
		// Include workdir context in the prompt for Ollama
		prompt = fmt.Sprintf("Current working directory: %s\n\n%s", workDir, prompt)
	}

	return o.sendToOllama(prompt, writer)
}

// sendToOllama makes the actual HTTP request to Ollama's /api/generate endpoint
func (o *OllamaClient) sendToOllama(prompt string, writer io.Writer) (string, error) {
	// Prepare request body
	reqBody := fmt.Sprintf(`{"model":"%s","prompt":"%s","stream":true,"raw":true}`,
		o.Model, escapeJSON(prompt))

	// Create HTTP request
	url := fmt.Sprintf("%s/api/generate", strings.TrimSuffix(o.BaseURL, "/"))
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama at %s: %w. Make sure Ollama is running with `ollama serve`", o.BaseURL, err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Stream the response
	var fullResponse bytes.Buffer
	buf := make([]byte, 4096)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			// Write to response writer if provided
			if writer != nil {
				if _, writeErr := writer.Write(chunk); writeErr != nil {
					return "", fmt.Errorf("failed to write response chunk: %w", writeErr)
				}
			}
			// Also accumulate for return value
			fullResponse.Write(chunk)
		}

		if err != nil {
			if err != io.EOF {
				return fullResponse.String(), fmt.Errorf("failed to read from ollama output: %w", err)
			}
			break
		}
	}

	return fullResponse.String(), nil
}

// escapeJSON escapes special characters for JSON string
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
