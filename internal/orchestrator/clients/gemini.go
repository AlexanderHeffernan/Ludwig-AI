package clients

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type GeminiClient struct{}

func (g *GeminiClient) SendPrompt(prompt string) (string, error) {
	cmd := exec.Command("gemini", "--yolo", prompt)
	var out bytes.Buffer
	var stderror bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderror

	err := cmd.Run()
	if err != nil {
		stderr := stderror.String()
		if stderr != "" {
			return "", fmt.Errorf("gemini command failed: %w\nstderr: %s", err, stderr)
		}
		return "", fmt.Errorf("gemini command failed: %w", err)
	}
	return out.String(), nil
}

// SendPromptWithStream sends a prompt and streams the response to the provided writer in real-time
// Uses --output-format stream-json to get all intermediate thoughts and reasoning as they happen
func (g *GeminiClient) SendPromptWithStream(prompt string, writer io.Writer) (string, error) {
	// Use --output-format stream-json for real-time event streaming
	// This outputs each thought, message, and tool call as newline-delimited JSON
	// allowing us to see the AI's complete reasoning process as it happens
	cmd := exec.Command("gemini", "--yolo", "--output-format", "stream-json", prompt)
	
	// Create a pipe to read stdout in real-time
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr separately for error reporting
	var stderror bytes.Buffer
	cmd.Stderr = &stderror

	// Start the command (don't wait yet)
	if err := cmd.Start(); err != nil {
		stderr := stderror.String()
		if stderr != "" {
			return "", fmt.Errorf("failed to start gemini: %w\nstderr: %s", err, stderr)
		}
		return "", fmt.Errorf("failed to start gemini: %w", err)
	}

	// Stream the output to the writer in real-time
	var fullResponse bytes.Buffer
	buf := make([]byte, 4096) // 4KB buffer for streaming chunks

	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			// Write to response writer (streams to file)
			if _, writeErr := writer.Write(chunk); writeErr != nil {
				cmd.Wait() // Clean up process
				return "", fmt.Errorf("failed to write response chunk: %w", writeErr)
			}
			// Also accumulate for return value
			fullResponse.Write(chunk)
		}

		if err != nil {
			if err != io.EOF {
				cmd.Wait() // Clean up process
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
