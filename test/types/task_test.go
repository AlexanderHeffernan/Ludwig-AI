package types_test

import (
	"testing"

	"ludwig/internal/types/task"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   task.Status
		expected string
	}{
		{
			name:     "Pending status",
			status:   task.Pending,
			expected: "Pending",
		},
		{
			name:     "InProgress status",
			status:   task.InProgress,
			expected: "In Progress",
		},
		{
			name:     "NeedsReview_status",
			status:   task.NeedsReview,
			expected: "In Review",
		},
		{
			name:     "Completed status",
			status:   task.Completed,
			expected: "Completed",
		},
		{
			name:     "Invalid status",
			status:   task.Status(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTask := task.Task{Status: tt.status}
			result := task.StatusString(testTask)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExampleTasks(t *testing.T) {
	tasks := task.ExampleTasks()

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

	for i, taskItem := range tasks {
		if taskItem.Status != task.Pending {
			t.Errorf("task %d: expected status Pending, got %v", i, taskItem.Status)
		}
	}
}

func TestTaskCreation(t *testing.T) {
	testTask := task.Task{
		ID:     "test-1",
		Name:   "Test Task",
		Status: task.InProgress,
	}

	if testTask.ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", testTask.ID)
	}
	if testTask.Name != "Test Task" {
		t.Errorf("expected name 'Test Task', got %s", testTask.Name)
	}
	if testTask.Status != task.InProgress {
		t.Errorf("expected status InProgress, got %v", testTask.Status)
	}
}

func TestReviewRequest(t *testing.T) {
	req := task.ReviewRequest{
		Question: "Should we refactor this?",
		Options: []task.ReviewOption{
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
	testTask := task.Task{
		ID:           "test-1",
		Name:         "Test Task",
		Status:       task.InProgress,
		BranchName:   "ludwig/test-task",
		WorktreePath: "/repo/.worktrees/test-1",
	}

	if testTask.WorktreePath != "/repo/.worktrees/test-1" {
		t.Errorf("expected worktree path /repo/.worktrees/test-1, got %s", testTask.WorktreePath)
	}
	if testTask.BranchName != "ludwig/test-task" {
		t.Errorf("expected branch name ludwig/test-task, got %s", testTask.BranchName)
	}
}
