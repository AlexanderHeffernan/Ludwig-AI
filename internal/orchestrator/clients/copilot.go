package clients

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type CopilotClient struct {
	Model string // e.g., "gpt-5" (default), "gpt-5-mini", "claude-sonnet-4.5"
}

// NewCopilotClient creates a new Copilot client with default settings
// Model defaults to gpt-5 (good balance of capability and speed)
func NewCopilotClient(model string) *CopilotClient {
	if model == "" {
		model = "gpt-5"
	}
	return &CopilotClient{
		Model: model,
	}
}

// SendPrompt sends a prompt to GitHub Copilot CLI with streaming
// - Streams output in real-time to the provided writer
// - Returns the complete response text once done
// - Runs in the current working directory (main repo)
func (c *CopilotClient) SendPrompt(prompt string, writer io.Writer) (string, error) {
	return c.SendPromptWithDir(prompt, writer, "")
}

// SendPromptWithDir sends a prompt to GitHub Copilot CLI in a specific working directory (e.g., worktree)
// - Same behavior as SendPrompt but executes in the provided workDir
// - If workDir is empty, uses current working directory
// - GitHub Copilot CLI runs with context awareness of the current directory
func (c *CopilotClient) SendPromptWithDir(prompt string, writer io.Writer, workDir string) (string, error) {
	return c.executeStreamInDir(prompt, writer, workDir)
}

// executeStreamInDir executes a single streaming request to Copilot in a specific working directory
// - Uses "copilot -p" for non-interactive mode with --allow-all-tools for automation
// - If workDir is empty, uses current working directory
func (c *CopilotClient) executeStreamInDir(prompt string, writer io.Writer, workDir string) (string, error) {
	// GitHub Copilot CLI command: copilot --model <model> -p <prompt> --allow-all-tools
	// --allow-all-tools is required for non-interactive/automated use
	cmd := exec.Command("copilot", "--model", c.Model, "-p", prompt, "--allow-all-tools")
	
	// Set working directory for the command if provided
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
			return "", fmt.Errorf("failed to start copilot: %w\nstderr: %s", err, stderr)
		}
		return "", fmt.Errorf("failed to start copilot: %w", err)
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
				return "", fmt.Errorf("failed to read from copilot output: %w", err)
			}
			break
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		stderr := stderror.String()
		response := fullResponse.String()
		if stderr != "" {
			return response, fmt.Errorf("copilot command exited with error: %w\nstderr: %s", err, stderr)
		}
		return response, fmt.Errorf("copilot command exited with error: %w", err)
	}

	return fullResponse.String(), nil
}
