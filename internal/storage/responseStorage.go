package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const ludwigDir = ".ludwig"

// getLudwigDirPath returns the path to the .ludwig directory within the current working directory.
func getLudwigDirPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	ludwigPath := filepath.Join(cwd, ludwigDir)
	return ludwigPath, nil
}

// ResponseWriter streams AI responses to a file
type ResponseWriter struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
	taskID   string
}

// NewResponseWriter creates a new response writer for a task
// Returns the relative path to the response file for storage in tasks.json
func NewResponseWriter(taskID string) (*ResponseWriter, string, error) {
	ludwigPath, err := getLudwigDirPath()
	if err != nil {
		return nil, "", err
	}

	// Create responses directory
	responseDir := filepath.Join(ludwigPath, "responses")
	if err := os.MkdirAll(responseDir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create .ludwig/responses directory: %w", err)
	}

	// Create filename with timestamp to ensure uniqueness
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.md", taskID, timestamp)
	filePath := filepath.Join(responseDir, filename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, "", err
	}

	// Write header
	header := fmt.Sprintf("# AI Response for Task: %s\n\nGenerated: %s\n\n---\n\n", taskID, time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(header); err != nil {
		file.Close()
		return nil, "", err
	}

	rw := &ResponseWriter{
		filePath: filePath,
		file:     file,
		taskID:   taskID,
	}

	// Return relative path for storage
	relativePath := filepath.Join("responses", filename) // This relative path is relative to .ludwig
	return rw, relativePath, nil
}

// WriteChunk writes a chunk of response data (streaming)
func (rw *ResponseWriter) WriteChunk(chunk string) error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.file == nil {
		return fmt.Errorf("response writer for task %s is closed", rw.taskID)
	}

	_, err := rw.file.WriteString(chunk)
	if err != nil {
		return err
	}

	// Flush to ensure data is written
	if err := rw.file.Sync(); err != nil {
		return err
	}

	return nil
}

// Write implements the io.Writer interface for compatibility
func (rw *ResponseWriter) Write(p []byte) (n int, err error) {
	if err := rw.WriteChunk(string(p)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close closes the response file
func (rw *ResponseWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.file == nil {
		return nil
	}

	// Write footer
	footer := fmt.Sprintf("\n\n---\n\nCompleted: %s\n", time.Now().Format(time.RFC3339))
	if _, err := rw.file.WriteString(footer); err != nil {
		rw.file.Close()
		rw.file = nil
		return err
	}

	// Ensure footer is synced to disk
	if err := rw.file.Sync(); err != nil {
		rw.file.Close()
		rw.file = nil
		return err
	}

	err := rw.file.Close()
	rw.file = nil
	return err
}

// GetFilePath returns the full file path
func (rw *ResponseWriter) GetFilePath() string {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.filePath
}

// ReadResponse reads the full response from file
func ReadResponse(filePath string) (string, error) {
	ludwigPath, err := getLudwigDirPath()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(ludwigPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
