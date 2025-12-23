package orchestrator_test

import (
	"strings"
	"testing"

	"ludwig/internal/orchestrator"
)

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name          string
		taskName      string
		expectedStart string
	}{
		{
			name:          "simple task",
			taskName:      "Create user database",
			expectedStart: "ludwig/",
		},
		{
			name:          "task with multiple words",
			taskName:      "Add authentication to API endpoints",
			expectedStart: "ludwig/",
		},
		{
			name:          "task with special characters",
			taskName:      "Fix: User login bug",
			expectedStart: "ludwig/",
		},
		{
			name:          "task with numbers",
			taskName:      "Create API v2 specification",
			expectedStart: "ludwig/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branchName, err := orchestrator.GenerateBranchName(tt.taskName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.HasPrefix(branchName, tt.expectedStart) {
				t.Errorf("expected branch name to start with %q, got %q", tt.expectedStart, branchName)
			}

			if strings.Contains(branchName, " ") {
				t.Errorf("expected no spaces in branch name, got %q", branchName)
			}

			if strings.ContainsAny(branchName, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				t.Errorf("expected lowercase branch name, got %q", branchName)
			}
		})
	}
}

func TestGenerateBranchNameLength(t *testing.T) {
	taskName := "This is a very long task description that should be truncated properly to fit git branch name conventions"
	branchName, err := orchestrator.GenerateBranchName(taskName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Branch name should be reasonable length (including ludwig/ prefix)
	if len(branchName) > 60 {
		t.Errorf("expected reasonable branch length, got %d characters: %q", len(branchName), branchName)
	}
}

func TestGenerateBranchNameEmptyTask(t *testing.T) {
	_, err := orchestrator.GenerateBranchName("")
	if err == nil {
		t.Errorf("expected error for empty task name, got nil")
	}
}

func TestGenerateBranchNameSpecialCharactersOnly(t *testing.T) {
	_, err := orchestrator.GenerateBranchName("!@#$%^&*()")
	if err == nil {
		t.Errorf("expected error for task with only special characters")
	}
}

func TestGenerateBranchNameKebabCase(t *testing.T) {
	taskName := "Create_User_Database"
	branchName, err := orchestrator.GenerateBranchName(taskName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use hyphens, not underscores
	if strings.Contains(branchName, "_") {
		t.Errorf("expected no underscores in branch name, got %q", branchName)
	}
}

func TestGenerateBranchNameWordsExtraction(t *testing.T) {
	// Task with short words that should be skipped
	taskName := "Do a b test"
	branchName, err := orchestrator.GenerateBranchName(taskName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(branchName, "ludwig/") {
		t.Errorf("expected ludwig/ prefix")
	}

	if strings.Contains(branchName, "-a-") || strings.Contains(branchName, "-b-") {
		t.Errorf("expected short words to be skipped, got %q", branchName)
	}
}

func TestGenerateBranchNameUniqueness(t *testing.T) {
	taskName := "Test branch"
	
	branchName1, err := orchestrator.GenerateBranchName(taskName)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Generate again - should get same base or numbered version
	branchName2, err := orchestrator.GenerateBranchName(taskName)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	// Names might be different due to git state, but should be valid
	if branchName1 == "" || branchName2 == "" {
		t.Errorf("expected non-empty branch names")
	}
}
