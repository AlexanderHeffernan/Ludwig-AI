package clients

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type GeminiClient struct{}

// modelFallbackChain defines the order in which models are tried
var modelFallbackChain = []string{
	"auto-gemini-3",
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
}

// SendPrompt sends a prompt to Gemini with streaming, retries on rate limits, and model fallback.
// - Tries models in order: auto-gemini-3, gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-lite
// - For each model, retries up to 3 times on rate limit (429) errors with exponential backoff
// - Streams output in real-time to the provided writer
// - On failure (non-rate-limit), falls back to the next weaker model
// - Returns the complete response text once done
// - Runs in the current working directory (main repo)
func (g *GeminiClient) SendPrompt(prompt string, writer io.Writer) (string, error) {
	return g.SendPromptWithDir(prompt, writer, "")
}

// SendPromptWithDir sends a prompt to Gemini in a specific working directory (e.g., worktree).
// - Same behavior as SendPrompt but executes in the provided workDir
// - If workDir is empty, uses current working directory
func (g *GeminiClient) SendPromptWithDir(prompt string, writer io.Writer, workDir string) (string, error) {
	for _, model := range modelFallbackChain {
		response, err := g.SendPromptWithModelAndDir(prompt, writer, model, workDir)
		
		// If successful, return
		if err == nil {
			return response, nil
		}
		
		// If it's a rate limit error, don't fall back - return immediately
		if isRateLimitError(response, err) {
			return response, err
		}
		
		// Non-rate-limit error: try next model
		msg := fmt.Sprintf("\n\n⚠️  Model %s failed: %v. Falling back to next model...\n\n", model, err)
		if writer != nil {
			writer.Write([]byte(msg))
		}
	}
	
	return "", fmt.Errorf("all models exhausted")
}

// SendPromptWithModel sends a prompt to Gemini using a specific model with rate limit retries
// - Retries up to 3 times on rate limit (429) errors with exponential backoff
// - Includes partial work from previous attempt so AI can catch up and continue
// - Returns the complete response text once done
// - Runs in the current working directory (main repo)
func (g *GeminiClient) SendPromptWithModel(prompt string, writer io.Writer, model string) (string, error) {
	return g.SendPromptWithModelAndDir(prompt, writer, model, "")
}

// SendPromptWithModelAndDir sends a prompt to Gemini in a specific directory using a specific model
// - Same behavior as SendPromptWithModel but executes in the provided workDir
// - If workDir is empty, uses current working directory
func (g *GeminiClient) SendPromptWithModelAndDir(prompt string, writer io.Writer, model string, workDir string) (string, error) {
	maxRetries := 3
	baseDelay := 30 * time.Second
	var lastPartialResponse string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// On retry, include previous partial work as context
		promptToUse := prompt
		if attempt > 0 && lastPartialResponse != "" {
			promptToUse = buildRetryPrompt(prompt, lastPartialResponse)
		}

		response, err := g.executeStreamInDir(promptToUse, writer, model, workDir)

		// Check for rate limit error (429)
		if isRateLimitError(response, err) {
			if attempt < maxRetries {
				lastPartialResponse = response // Save partial work for next attempt
				delay := baseDelay * time.Duration(1<<uint(attempt)) // 30s, 60s, 120s
				msg := fmt.Sprintf("\n\n⚠️  Rate limited. Retrying in %v... (attempt %d/%d)\n\n", delay, attempt+1, maxRetries)
				if writer != nil {
					writer.Write([]byte(msg))
				}
				time.Sleep(delay)
				continue
			}
			// Out of retries
			return response, fmt.Errorf("rate limit exceeded after %d retries: %w", maxRetries, err)
		}

		// Non-rate-limit error or success
		return response, err
	}

	return "", fmt.Errorf("max retries exceeded")
}

// executeStream executes a single streaming request to Gemini using a specific model
// - Runs in the current working directory (main repo)
func (g *GeminiClient) executeStream(prompt string, writer io.Writer, model string) (string, error) {
	return g.executeStreamInDir(prompt, writer, model, "")
}

// executeStreamInDir executes a single streaming request to Gemini in a specific working directory
// - If workDir is empty, uses current working directory
func (g *GeminiClient) executeStreamInDir(prompt string, writer io.Writer, model string, workDir string) (string, error) {
	// Use --output-format stream-json for real-time event streaming
	cmd := exec.Command("gemini", "--yolo", "--model", model, "--output-format", "stream-json", prompt)
	
	// Set working directory for the command
	if workDir != "" {
		cmd.Dir = workDir
	}

	// Create a pipe to read stdout in real-time
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr separately for error reporting
	var stderror bytes.Buffer
	cmd.Stderr = &stderror

	// Start the command
	if err := cmd.Start(); err != nil {
		stderr := stderror.String()
		if stderr != "" {
			return "", fmt.Errorf("failed to start gemini: %w\nstderr: %s", err, stderr)
		}
		return "", fmt.Errorf("failed to start gemini: %w", err)
	}

	// Stream the output to the writer in real-time
	var fullResponse bytes.Buffer
	buf := make([]byte, 4096)

	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			// Write to response writer (streams to file immediately)
			if _, writeErr := writer.Write(chunk); writeErr != nil {
				cmd.Wait()
				return "", fmt.Errorf("failed to write response chunk: %w", writeErr)
			}
			// Also accumulate for return value
			fullResponse.Write(chunk)
		}

		if err != nil {
			if err != io.EOF {
				cmd.Wait()
				return "", fmt.Errorf("failed to read from gemini output: %w", err)
			}
			break
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		stderr := stderror.String()
		response := fullResponse.String()
		if stderr != "" {
			return response, fmt.Errorf("gemini command exited with error: %w\nstderr: %s", err, stderr)
		}
		return response, fmt.Errorf("gemini command exited with error: %w", err)
	}

	return fullResponse.String(), nil
}

// buildRetryPrompt creates a new prompt that includes the partial work from the previous attempt
// This allows the AI to catch up on what was already done and continue from where it left off
func buildRetryPrompt(originalPrompt string, partialResponse string) string {
	if partialResponse == "" {
		return originalPrompt
	}

	return fmt.Sprintf(`%s

---

[PREVIOUS WORK COMPLETED ON RETRY]:
%s
[END PREVIOUS WORK]

Please review the above work. If it appears complete, confirm that and provide a summary. If it's incomplete, continue from where it left off to finish the task.`,
		originalPrompt, partialResponse)
}

// isRateLimitError checks if the error is a 429 rate limit error
func isRateLimitError(response string, err error) bool {
	// Check response for rate limit indicators
	if response != "" {
		lowerResponse := strings.ToLower(response)
		if strings.Contains(lowerResponse, "resource has been exhausted") ||
			strings.Contains(lowerResponse, "429") ||
			strings.Contains(lowerResponse, "rate limit") ||
			strings.Contains(lowerResponse, "too many requests") {
			return true
		}
	}

	// Check error message
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "resource has been exhausted") ||
			strings.Contains(lowerErr, "429") ||
			strings.Contains(lowerErr, "rate limit") ||
			strings.Contains(lowerErr, "too many requests") {
			return true
		}
	}

	return false
}
