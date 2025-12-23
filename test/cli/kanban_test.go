package cli_test

import (
	"strings"
	"testing"

	"ludwig/internal/cli"
)

func TestKanbanTaskNameNormal(t *testing.T) {
	name := "User Login"
	result := cli.KanbanTaskName(name)

	if !strings.Contains(result, name) {
		t.Errorf("expected result to contain task name %q, got %q", name, result)
	}

	if !strings.Contains(result, "│") {
		t.Errorf("expected result to have kanban border character")
	}
}

func TestKanbanTaskNameEmpty(t *testing.T) {
	result := cli.KanbanTaskName("")

	// Should still produce a valid kanban cell
	if !strings.Contains(result, "│") {
		t.Errorf("expected empty task to produce kanban cell")
	}
}

func TestKanbanTaskNameTruncation(t *testing.T) {
	longName := "This is an extremely long task name that definitely should be truncated"
	result := cli.KanbanTaskName(longName)

	// Should contain truncation indicator
	if !strings.Contains(result, "...") {
		t.Errorf("expected truncated name to contain '...', got %q", result)
	}

	// Should be reasonable length
	if len(result) > 50 {
		t.Errorf("expected reasonable length after truncation, got %d chars: %q", len(result), result)
	}
}

// Note: seperateTaskByStatus is private, so we can only test public APIs
// The public API tested is KanbanTaskName()

func TestKanbanTaskNameBoundary(t *testing.T) {
	// Test with name exactly at boundary
	name := "Task1234567890123456"
	result := cli.KanbanTaskName(name)

	if !strings.Contains(result, "│") {
		t.Errorf("expected kanban cell format")
	}
}

func TestKanbanTaskNameSpecialChars(t *testing.T) {
	name := "Task: Feature #123"
	result := cli.KanbanTaskName(name)

	if !strings.Contains(result, "│") {
		t.Errorf("expected kanban cell format with special chars")
	}

	if !strings.Contains(result, "Task") {
		t.Errorf("expected name to be preserved")
	}
}

func TestKanbanTaskNameUnicode(t *testing.T) {
	name := "Task 📝"
	result := cli.KanbanTaskName(name)

	if !strings.Contains(result, "│") {
		t.Errorf("expected kanban cell format with unicode")
	}
}
