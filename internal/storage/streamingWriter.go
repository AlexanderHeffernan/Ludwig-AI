package storage

import (
	"fmt"
)

// StreamingWriter wraps ResponseWriter to implement io.Writer interface
type StreamingWriter struct {
	resp *ResponseWriter
}

// NewStreamingWriter creates a streaming writer that wraps a ResponseWriter
func NewStreamingWriter(respWriter *ResponseWriter) *StreamingWriter {
	return &StreamingWriter{resp: respWriter}
}

// Write implements the io.Writer interface
func (sw *StreamingWriter) Write(p []byte) (n int, err error) {
	if err := sw.resp.WriteChunk(string(p)); err != nil {
		return 0, fmt.Errorf("failed to write chunk: %w", err)
	}
	return len(p), nil
}
