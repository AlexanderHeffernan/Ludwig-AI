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

// SendPrompt sends a prompt to Gemini with streaming, retries on rate limits, and returns the response
// - Streams output in real-time to the provided writer
// - Automatically retries with exponential backoff on rate limit (429) errors
// - On retry, includes partial work from previous attempt so AI can catch up and continue
// - Returns the complete response text once done
func (g *GeminiClient) SendPrompt(prompt string, writer io.Writer) (string, error) {
	maxRetries := 3
	baseDelay := 30 * time.Second
	var lastPartialResponse string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// On retry, include previous partial work as context
		promptToUse := prompt
		if attempt > 0 && lastPartialResponse != "" {
			promptToUse = buildRetryPrompt(prompt, lastPartialResponse)
		}

		response, err := g.executeStream(promptToUse, writer)

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

// executeStream executes a single streaming request to Gemini
func (g *GeminiClient) executeStream(prompt string, writer io.Writer) (string, error) {
	// Use --output-format stream-json for real-time event streaming
	cmd := exec.Command("gemini", "--yolo", "--output-format", "stream-json", prompt)

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
