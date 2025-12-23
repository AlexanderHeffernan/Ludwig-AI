package storage_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ludwig/internal/storage"
)

func cleanupResponseStorage(t *testing.T) {
	home, _ := os.UserHomeDir()
	responseDir := filepath.Join(home, ".ai-orchestrator", "responses")
	os.RemoveAll(responseDir)
}

func TestNewResponseWriter(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, relativePath, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	if rw == nil {
		t.Errorf("expected response writer, got nil")
	}
	if relativePath == "" {
		t.Errorf("expected non-empty relative path, got empty")
	}
	if !strings.Contains(relativePath, "responses") {
		t.Errorf("expected path to contain 'responses', got %s", relativePath)
	}
	if !strings.Contains(relativePath, "test-task") {
		t.Errorf("expected path to contain task ID, got %s", relativePath)
	}
}

func TestResponseWriterWriteChunk(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	testContent := "This is test content"
	if err := rw.WriteChunk(testContent); err != nil {
		t.Fatalf("failed to write chunk: %v", err)
	}

	// Close and read the file to verify content
	filePath := rw.GetFilePath()
	rw.Close()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, testContent) {
		t.Errorf("expected content to contain %q, got %s", testContent, contentStr)
	}
}

func TestResponseWriterMultipleChunks(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	chunks := []string{"chunk1", "chunk2", "chunk3"}
	for _, chunk := range chunks {
		if err := rw.WriteChunk(chunk); err != nil {
			t.Fatalf("failed to write chunk: %v", err)
		}
	}

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	contentStr := string(content)

	for _, chunk := range chunks {
		if !strings.Contains(contentStr, chunk) {
			t.Errorf("expected content to contain %q", chunk)
		}
	}
}

func TestResponseWriterClose(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}

	if err := rw.Close(); err != nil {
		t.Errorf("failed to close response writer: %v", err)
	}

	// Second close should be idempotent
	if err := rw.Close(); err != nil {
		t.Errorf("expected no error on second close, got %v", err)
	}
}

func TestResponseWriterHeader(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "# AI Response for Task: test-task") {
		t.Errorf("expected header with task ID in content")
	}
	if !strings.Contains(contentStr, "Generated:") {
		t.Errorf("expected Generated timestamp in content")
	}
}

func TestResponseWriterFooter(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "Completed:") {
		t.Errorf("expected Completed timestamp in footer")
	}
}

func TestReadResponse(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, relativePath, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}

	testContent := "Test content to read"
	rw.WriteChunk(testContent)
	rw.Close()

	// Read using ReadResponse function
	content, err := storage.ReadResponse(relativePath)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if !strings.Contains(content, testContent) {
		t.Errorf("expected read content to contain %q, got %s", testContent, content)
	}
}

func TestResponseWriterWriteInterface(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	testData := []byte("test data through Write interface")
	n, err := rw.Write(testData)
	if err != nil {
		t.Fatalf("failed to write via interface: %v", err)
	}

	if n != len(testData) {
		t.Errorf("expected %d bytes written, got %d", len(testData), n)
	}

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	if !strings.Contains(string(content), string(testData)) {
		t.Errorf("expected written data in file")
	}
}
