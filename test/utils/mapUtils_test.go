package utils_test

import (
	"testing"

	"ludwig/internal/types/task"
	"ludwig/internal/utils"
)

func TestPointerSliceToValueSlice(t *testing.T) {
	// Test with nil input
	result := utils.PointerSliceToValueSlice(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	// Test with empty slice
	result = utils.PointerSliceToValueSlice([]*task.Task{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}

	// Test with valid tasks
	task1 := &task.Task{ID: "1"}
	task2 := &task.Task{ID: "2"}
	result = utils.PointerSliceToValueSlice([]*task.Task{task1, task2})
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
	}
	if result[0].ID != "1" {
		t.Errorf("expected task id '1', got %s", result[0].ID)
	}
}
