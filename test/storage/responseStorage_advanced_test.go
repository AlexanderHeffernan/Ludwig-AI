package storage_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ludwig/internal/storage"
)

// Test response file naming convention
func TestResponseFileNamingConvention(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, relativePath, _ := storage.NewResponseWriter("my-task-123")
	defer rw.Close()

	// Should contain task ID
	if !strings.Contains(relativePath, "my-task-123") {
		t.Errorf("relative path should contain task ID: %s", relativePath)
	}

	// Should contain responses directory
	if !strings.Contains(relativePath, "responses") {
		t.Errorf("relative path should contain responses directory: %s", relativePath)
	}

	// Should be a markdown file
	if !strings.HasSuffix(relativePath, ".md") {
		t.Errorf("relative path should end with .md: %s", relativePath)
	}
}

// Test response file with special characters in task ID
func TestResponseFileSpecialCharacters(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, relativePath, _ := storage.NewResponseWriter("task-with-special_chars")
	defer rw.Close()

	if !strings.Contains(relativePath, "task-with-special_chars") {
		t.Errorf("should handle task IDs with special characters")
	}
}

// Test response writer header and footer
func TestResponseWriterHeaderFooterFormat(t *testing.T) {
	defer cleanupResponseStorage(t)

	taskID := "test-task"
	rw, _, _ := storage.NewResponseWriter(taskID)
	rw.WriteChunk("Main content here")
	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	contentStr := string(content)

	// Check header
	if !strings.Contains(contentStr, "# AI Response for Task:") {
		t.Errorf("missing header")
	}
	if !strings.Contains(contentStr, taskID) {
		t.Errorf("header should contain task ID")
	}
	if !strings.Contains(contentStr, "Generated:") {
		t.Errorf("missing Generated timestamp")
	}

	// Check footer
	if !strings.Contains(contentStr, "Completed:") {
		t.Errorf("missing Completed timestamp")
	}

	// Check content order
	generatedIdx := strings.Index(contentStr, "Generated:")
	completedIdx := strings.Index(contentStr, "Completed:")
	if generatedIdx >= completedIdx {
		t.Errorf("Generated should come before Completed")
	}
}

// Test large response writing
func TestResponseWriterLargeContent(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("large-task")
	defer rw.Close()

	// Write large content
	largeContent := strings.Repeat("This is a test content line.\n", 1000)
	err := rw.WriteChunk(largeContent)
	if err != nil {
		t.Errorf("failed to write large content: %v", err)
	}

	filePath := rw.GetFilePath()
	rw.Close()

	// Verify file size
	info, _ := os.Stat(filePath)
	if info.Size() == 0 {
		t.Errorf("file should contain written content")
	}
}

// Test response writer with binary-like content
func TestResponseWriterBinaryContent(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("binary-task")
	defer rw.Close()

	// Write content with special bytes
	content := "Text with special chars: \n\t\r special symbols !@#$%"
	err := rw.WriteChunk(content)
	if err != nil {
		t.Errorf("failed to write special content: %v", err)
	}

	filePath := rw.GetFilePath()
	rw.Close()

	readContent, _ := os.ReadFile(filePath)
	if !strings.Contains(string(readContent), "special chars") {
		t.Errorf("special content not preserved")
	}
}

// Test GetFilePath consistency
func TestResponseWriterFilePathConsistency(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("consistency-test")

	path1 := rw.GetFilePath()
	path2 := rw.GetFilePath()

	if path1 != path2 {
		t.Errorf("GetFilePath should return consistent path")
	}

	rw.Close()
}

// Test writing after close
func TestResponseWriterWriteAfterClose(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("close-test")
	rw.Close()

	// Writing after close should error
	err := rw.WriteChunk("This should fail")
	if err == nil {
		t.Errorf("should error when writing after close")
	}
}

// Test ReadResponse with non-existent file
func TestReadResponseNonExistent(t *testing.T) {
	_, err := storage.ReadResponse("responses/nonexistent-file.md")
	if err == nil {
		t.Errorf("should error when reading non-existent response")
	}
}

// Test response writer idempotent close
func TestResponseWriterIdempotentClose(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("idempotent-test")
	rw.WriteChunk("content")

	// Multiple closes should be safe
	err1 := rw.Close()
	err2 := rw.Close()
	err3 := rw.Close()

	if err1 != nil {
		t.Errorf("first close should succeed")
	}
	// Subsequent closes should be safe (either error or succeed)
	if err2 != nil && !strings.Contains(err2.Error(), "closed") {
		t.Errorf("second close should be safe, got: %v", err2)
	}
	if err3 != nil && !strings.Contains(err3.Error(), "closed") {
		t.Errorf("third close should be safe, got: %v", err3)
	}
}

// Test response file directory structure
func TestResponseFileDirectoryStructure(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("dir-test")
	defer rw.Close()

	filePath := rw.GetFilePath()

	// File should exist
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("response file should exist: %v", err)
	}

	// Parent directory should exist
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("response directory should exist: %v", err)
	}
}

// Test WriteChunk error handling
func TestWriteChunkErrorHandling(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("error-test")

	// Normal write should succeed
	err := rw.WriteChunk("test")
	if err != nil {
		t.Errorf("normal write should succeed: %v", err)
	}

	rw.Close()

	// Write after close should fail
	err = rw.WriteChunk("after close")
	if err == nil {
		t.Errorf("write after close should error")
	}
}

// Test response writer with timestamp
func TestResponseWriterTimestamp(t *testing.T) {
	defer cleanupResponseStorage(t)

	before := time.Now()
	rw, _, _ := storage.NewResponseWriter("timestamp-test")
	_ = time.Now()

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	contentStr := string(content)

	// Should contain timestamps
	if !strings.Contains(contentStr, "Generated:") {
		t.Errorf("should have Generated timestamp")
	}
	if !strings.Contains(contentStr, "Completed:") {
		t.Errorf("should have Completed timestamp")
	}

	// Timestamps should be reasonable
	if !strings.Contains(contentStr, before.Format("2006")) {
		t.Logf("timestamp format check (may vary): before=%s", before)
	}
}

// Test response writer with multiple task IDs
func TestResponseWriterMultipleTaskIDs(t *testing.T) {
	defer cleanupResponseStorage(t)

	taskIDs := []string{"task-1", "task-2", "task-3"}
	filePaths := []string{}

	for _, taskID := range taskIDs {
		rw, _, _ := storage.NewResponseWriter(taskID)
		rw.WriteChunk("Content for " + taskID)
		filePaths = append(filePaths, rw.GetFilePath())
		rw.Close()
	}

	// All files should exist and be different
	if len(filePaths) != len(taskIDs) {
		t.Errorf("should have created all files")
	}

	for i, path := range filePaths {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("task %d file should exist: %s", i, path)
		}
	}

	// Files should be different
	if filePaths[0] == filePaths[1] {
		t.Errorf("different tasks should have different response files")
	}
}

// Test streaming interface compliance
func TestResponseWriterStreamingInterface(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, _ := storage.NewResponseWriter("streaming-test")
	defer rw.Close()

	// Test Write interface (io.Writer compatible)
	data := []byte("test data")
	n, err := rw.Write(data)

	if err != nil {
		t.Errorf("Write should not error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write should return bytes written: got %d, expected %d", n, len(data))
	}
}

// Test ReadResponse with valid file
func TestReadResponseValidFile(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, relativePath, _ := storage.NewResponseWriter("read-test")
	expectedContent := "This is the response content"
	rw.WriteChunk(expectedContent)
	rw.Close()

	content, err := storage.ReadResponse(relativePath)
	if err != nil {
		t.Errorf("ReadResponse should succeed: %v", err)
	}

	if !strings.Contains(content, expectedContent) {
		t.Errorf("ReadResponse should contain written content")
	}
}
