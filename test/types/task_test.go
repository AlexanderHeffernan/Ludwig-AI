package types_test

import (
	"testing"

	"ludwig/internal/types"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   types.Status
		expected string
	}{
		{
			name:     "Pending status",
			status:   types.Pending,
			expected: "Pending",
		},
		{
			name:     "InProgress status",
			status:   types.InProgress,
			expected: "In Progress",
		},
		{
			name:     "NeedsReview_status",
			status:   types.NeedsReview,
			expected: "In Review",
		},
		{
			name:     "Completed status",
			status:   types.Completed,
			expected: "Completed",
		},
		{
			name:     "Invalid status",
			status:   types.Status(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := types.Task{Status: tt.status}
			result := types.StatusString(task)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExampleTasks(t *testing.T) {
	tasks := types.ExampleTasks()

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}

	expectedIDs := []string{"task-1", "task-2", "task-3"}
	for i, expectedID := range expectedIDs {
		if tasks[i].ID != expectedID {
			t.Errorf("task %d: expected ID %q, got %q", i, expectedID, tasks[i].ID)
		}
	}

	expectedNames := []string{"Create user authentication", "Setup database schema", "Design API endpoints"}
	for i, expectedName := range expectedNames {
		if tasks[i].Name != expectedName {
			t.Errorf("task %d: expected name %q, got %q", i, expectedName, tasks[i].Name)
		}
	}

	for i, task := range tasks {
		if task.Status != types.Pending {
			t.Errorf("task %d: expected status Pending, got %v", i, task.Status)
		}
	}
}

func TestTaskCreation(t *testing.T) {
	task := types.Task{
		ID:     "test-1",
		Name:   "Test Task",
		Status: types.InProgress,
	}

	if task.ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", task.ID)
	}
	if task.Name != "Test Task" {
		t.Errorf("expected name 'Test Task', got %s", task.Name)
	}
	if task.Status != types.InProgress {
		t.Errorf("expected status InProgress, got %v", task.Status)
	}
}

func TestReviewRequest(t *testing.T) {
	req := types.ReviewRequest{
		Question: "Should we refactor this?",
		Options: []types.ReviewOption{
			{ID: "yes", Label: "Yes, proceed"},
			{ID: "no", Label: "No, keep as is"},
		},
	}

	if req.Question != "Should we refactor this?" {
		t.Errorf("expected question, got %s", req.Question)
	}
	if len(req.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(req.Options))
	}
}

func TestTaskWithWorktreePath(t *testing.T) {
	task := types.Task{
		ID:           "test-1",
		Name:         "Test Task",
		Status:       types.InProgress,
		BranchName:   "ludwig/test-task",
		WorktreePath: "/repo/.worktrees/test-1",
	}

	if task.WorktreePath != "/repo/.worktrees/test-1" {
		t.Errorf("expected worktree path /repo/.worktrees/test-1, got %s", task.WorktreePath)
	}
	if task.BranchName != "ludwig/test-task" {
		t.Errorf("expected branch name ludwig/test-task, got %s", task.BranchName)
	}
}
