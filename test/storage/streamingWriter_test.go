package storage_test

import (
	"os"
	"strings"
	"testing"

	"ludwig/internal/storage"
)

func TestNewStreamingWriter(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	sw := storage.NewStreamingWriter(rw)
	if sw == nil {
		t.Errorf("expected streaming writer, got nil")
	}
}

func TestStreamingWriterWrite(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	sw := storage.NewStreamingWriter(rw)

	testData := []byte("streamed content")
	n, err := sw.Write(testData)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if n != len(testData) {
		t.Errorf("expected %d bytes written, got %d", len(testData), n)
	}

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	if !strings.Contains(string(content), "streamed content") {
		t.Errorf("expected streamed content in file")
	}
}

func TestStreamingWriterMultipleWrites(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	sw := storage.NewStreamingWriter(rw)

	writes := [][]byte{
		[]byte("first line\n"),
		[]byte("second line\n"),
		[]byte("third line\n"),
	}

	for _, data := range writes {
		_, err := sw.Write(data)
		if err != nil {
			t.Fatalf("failed to write: %v", err)
		}
	}

	filePath := rw.GetFilePath()
	rw.Close()

	content, _ := os.ReadFile(filePath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "first line") {
		t.Errorf("expected first line in file")
	}
	if !strings.Contains(contentStr, "second line") {
		t.Errorf("expected second line in file")
	}
	if !strings.Contains(contentStr, "third line") {
		t.Errorf("expected third line in file")
	}
}

func TestStreamingWriterEmptyWrite(t *testing.T) {
	defer cleanupResponseStorage(t)

	rw, _, err := storage.NewResponseWriter("test-task")
	if err != nil {
		t.Fatalf("failed to create response writer: %v", err)
	}
	defer rw.Close()

	sw := storage.NewStreamingWriter(rw)

	n, err := sw.Write([]byte{})
	if err != nil {
		t.Fatalf("failed to write empty: %v", err)
	}

	if n != 0 {
		t.Errorf("expected 0 bytes written, got %d", n)
	}
}
