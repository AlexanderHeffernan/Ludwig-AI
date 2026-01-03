package orchestrator_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ludwig/internal/orchestrator/clients"
)

// TestOllamaClientInitialization tests that Ollama client is created with correct defaults
func TestOllamaClientNewWithDefaults(t *testing.T) {
	client := clients.NewOllamaClient("", "")
	
	if client.BaseURL != "http://localhost:11434" {
		t.Errorf("expected default BaseURL to be 'http://localhost:11434', got '%s'", client.BaseURL)
	}
	
	if client.Model != "mistral" {
		t.Errorf("expected default Model to be 'mistral', got '%s'", client.Model)
	}
}

// TestOllamaClientNewWithCustomValues tests that Ollama client is created with custom values
func TestOllamaClientNewWithCustomValues(t *testing.T) {
	customURL := "http://192.168.1.100:11434"
	customModel := "neural-chat"
	client := clients.NewOllamaClient(customURL, customModel)
	
	if client.BaseURL != customURL {
		t.Errorf("expected BaseURL to be '%s', got '%s'", customURL, client.BaseURL)
	}
	
	if client.Model != customModel {
		t.Errorf("expected Model to be '%s', got '%s'", customModel, client.Model)
	}
}

// TestOllamaClientSendPrompt tests that SendPrompt calls the Ollama API correctly
func TestOllamaClientSendPrompt(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		
		if !strings.Contains(r.URL.Path, "/api/generate") {
			t.Errorf("expected /api/generate endpoint, got %s", r.URL.Path)
		}
		
		// Return a simple streaming response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"model":"mistral","created_at":"2024-01-01T00:00:00Z","response":"Hello"}`))
		w.Write([]byte(`{"model":"mistral","created_at":"2024-01-01T00:00:01Z","response":" world"}`))
	}))
	defer server.Close()
	
	client := clients.NewOllamaClient(server.URL, "mistral")
	
	var output bytes.Buffer
	response, err := client.SendPrompt("test prompt", &output)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if !strings.Contains(response, "Hello") {
		t.Errorf("expected response to contain 'Hello', got '%s'", response)
	}
}

// TestOllamaClientSendPromptWithDir tests that working directory context is included
func TestOllamaClientSendPromptWithDir(t *testing.T) {
	receivedBody := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"model":"mistral","response":"ok"}`))
	}))
	defer server.Close()
	
	client := clients.NewOllamaClient(server.URL, "mistral")
	
	_, err := client.SendPromptWithDir("test prompt", nil, "/tmp/workdir")
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify that workdir context was included in the prompt
	if !strings.Contains(receivedBody, "working directory") && !strings.Contains(receivedBody, "/tmp/workdir") {
		t.Errorf("expected working directory context in prompt, got '%s'", receivedBody)
	}
}

// TestOllamaClientConnectionError tests handling of connection errors
func TestOllamaClientConnectionError(t *testing.T) {
	// Create a client pointing to a non-existent server
	client := clients.NewOllamaClient("http://localhost:9999", "mistral")
	
	_, err := client.SendPrompt("test prompt", nil)
	
	if err == nil {
		t.Errorf("expected error when connecting to non-existent Ollama server")
	}
	
	if !strings.Contains(err.Error(), "failed to connect to Ollama") && !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected connection error message, got: %v", err)
	}
}

// TestOllamaClientHTTPError tests handling of HTTP errors from Ollama
func TestOllamaClientHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()
	
	client := clients.NewOllamaClient(server.URL, "mistral")
	
	_, err := client.SendPrompt("test prompt", nil)
	
	if err == nil {
		t.Errorf("expected error for HTTP 500")
	}
	
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

// TestOllamaClientStreaming tests that response is streamed correctly to writer
func TestOllamaClientStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response "))
		w.Write([]byte("chunk 1 "))
		w.Write([]byte("chunk 2"))
	}))
	defer server.Close()
	
	client := clients.NewOllamaClient(server.URL, "mistral")
	
	var output bytes.Buffer
	response, err := client.SendPrompt("test prompt", &output)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify response
	if !strings.Contains(response, "Response") {
		t.Errorf("expected response to contain 'Response', got '%s'", response)
	}
	
	// Verify streaming to writer
	if !strings.Contains(output.String(), "Response") {
		t.Errorf("expected output buffer to contain 'Response', got '%s'", output.String())
	}
}

// TestOllamaClientAIClientInterface tests that OllamaClient implements AIClient interface
func TestOllamaClientAIClientInterface(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response":"ok"}`))
	}))
	defer server.Close()
	
	client := clients.NewOllamaClient(server.URL, "mistral")
	var aiClient clients.AIClient = client
	
	// Should compile and work without errors
	_, err := aiClient.SendPrompt("test", nil)
	if err != nil {
		t.Errorf("AIClient interface method failed: %v", err)
	}
	
	_, err = aiClient.SendPromptWithDir("test", nil, "/tmp")
	if err != nil {
		t.Errorf("AIClient interface method with dir failed: %v", err)
	}
}

// TestGeminiClientAIClientInterface ensures GeminiClient still implements AIClient interface
func TestGeminiClientAIClientInterface(t *testing.T) {
	client := &clients.GeminiClient{}
	var aiClient clients.AIClient = client
	
	// Just verify it compiles (we can't easily test execution without gemini CLI)
	if aiClient == nil {
		t.Errorf("AIClient should not be nil")
	}
}

// TestCopilotClientNewWithDefaults tests that Copilot client is created with correct defaults
func TestCopilotClientNewWithDefaults(t *testing.T) {
	client := clients.NewCopilotClient("")
	
	if client.Model != "gpt-5" {
		t.Errorf("expected default Model to be 'gpt-5', got '%s'", client.Model)
	}
}

// TestCopilotClientNewWithCustomModel tests that Copilot client is created with custom model
func TestCopilotClientNewWithCustomModel(t *testing.T) {
	customModel := "claude-sonnet-4.5"
	client := clients.NewCopilotClient(customModel)
	
	if client.Model != customModel {
		t.Errorf("expected Model to be '%s', got '%s'", customModel, client.Model)
	}
}

// TestCopilotClientAIClientInterface ensures CopilotClient implements AIClient interface
func TestCopilotClientAIClientInterface(t *testing.T) {
	client := clients.NewCopilotClient("")
	var aiClient clients.AIClient = client
	
	// Verify the interface is properly implemented (can't execute without copilot CLI)
	if aiClient == nil {
		t.Errorf("AIClient should not be nil")
	}
}
