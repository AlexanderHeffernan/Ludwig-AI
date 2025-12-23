package orchestrator_test

import (
	"testing"
	"time"

	"ludwig/internal/orchestrator"
	"ludwig/internal/types"
)

// Test orchestrator lifecycle
func TestOrchestratorLifecycle(t *testing.T) {
	// Clean up any previous state
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	// Test initial state
	if orchestrator.IsRunning() {
		t.Errorf("orchestrator should not be running initially")
	}

	// Start
	orchestrator.Start()
	if !orchestrator.IsRunning() {
		t.Errorf("orchestrator should be running after Start()")
	}

	// Stop
	orchestrator.Stop()
	if orchestrator.IsRunning() {
		t.Errorf("orchestrator should not be running after Stop()")
	}
}

// Test quick start-stop cycle
func TestOrchestratorQuickStartStop(t *testing.T) {
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	for i := 0; i < 3; i++ {
		orchestrator.Start()
		time.Sleep(10 * time.Millisecond)
		orchestrator.Stop()
	}

	if orchestrator.IsRunning() {
		t.Errorf("orchestrator should be stopped at end of cycle")
	}
}

// Test branch name generation from various task formats
func TestBranchNameGenerationVariants(t *testing.T) {
	testCases := []struct {
		taskName string
		shouldWork bool
	}{
		{"Create user database", true},
		{"Add feature X", true},
		{"Fix bug in login", true},
		{"Implement API v2", true},
		{"Add-hyphens-to-task", true},
		{"Task with many many many words that are quite long", true},
		{"", false},
		{"   ", false},
	}

	for _, tc := range testCases {
		name, err := orchestrator.GenerateBranchName(tc.taskName)
		
		if tc.shouldWork && err != nil {
			t.Errorf("GenerateBranchName(%q) failed unexpectedly: %v", tc.taskName, err)
		}
		
		if !tc.shouldWork && err == nil {
			t.Errorf("GenerateBranchName(%q) should have failed but got: %s", tc.taskName, name)
		}

		if tc.shouldWork && name == "" {
			t.Errorf("GenerateBranchName(%q) returned empty name", tc.taskName)
		}

		if tc.shouldWork && len(name) == 0 {
			t.Errorf("branch name should have content")
		}
	}
}

// Test task prompt building consistency
func TestPromptBuildingConsistency(t *testing.T) {
	taskName := "Implement authentication"

	// Build same prompt multiple times
	prompts := make([]string, 3)
	for i := 0; i < 3; i++ {
		prompts[i] = orchestrator.BuildTaskPrompt(taskName)
	}

	// All should be identical
	for i := 1; i < len(prompts); i++ {
		if prompts[i] != prompts[0] {
			t.Errorf("prompts should be consistent, but differ at index %d", i)
		}
	}
}

// Test resume prompt with different option counts
func TestResumePromptDifferentOptionCounts(t *testing.T) {
	testCases := []struct {
		name    string
		options []string
	}{
		{
			name:    "single option",
			options: []string{"Proceed"},
		},
		{
			name:    "two options",
			options: []string{"Yes", "No"},
		},
		{
			name:    "three options",
			options: []string{"Option A", "Option B", "Option C"},
		},
		{
			name:    "many options",
			options: []string{"A", "B", "C", "D", "E", "F"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prompt := orchestrator.BuildResumePrompt(
				"Task",
				"Work",
				"Question?",
				tc.options,
				tc.options[0],
				"",
			)

			if prompt == "" {
				t.Errorf("prompt should not be empty")
			}

			for _, opt := range tc.options {
				if opt == "" {
					continue
				}
				if !containsString(prompt, opt) {
					t.Errorf("prompt should contain option %q", opt)
				}
			}
		})
	}
}

// Test orchestrator state isolation
func TestOrchestratorStateIsolation(t *testing.T) {
	// Ensure clean state
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	// Multiple Stop calls should not cause issues
	orchestrator.Stop()
	orchestrator.Stop()

	if orchestrator.IsRunning() {
		t.Errorf("orchestrator should not be running")
	}
}

// Test branch name format consistency
func TestBranchNameFormatConsistency(t *testing.T) {
	taskName := "Create user authentication system"

	name1, _ := orchestrator.GenerateBranchName(taskName)
	name2, _ := orchestrator.GenerateBranchName(taskName)

	// Both should have ludwig/ prefix
	if !hasPrefix(name1, "ludwig/") {
		t.Errorf("branch name missing ludwig/ prefix: %s", name1)
	}
	if !hasPrefix(name2, "ludwig/") {
		t.Errorf("branch name missing ludwig/ prefix: %s", name2)
	}

	// Neither should have spaces
	if containsString(name1, " ") {
		t.Errorf("branch name should not have spaces: %s", name1)
	}
	if containsString(name2, " ") {
		t.Errorf("branch name should not have spaces: %s", name2)
	}

	// Should not have uppercase
	for _, ch := range name1 {
		if ch >= 'A' && ch <= 'Z' {
			t.Errorf("branch name should be lowercase: %s", name1)
			break
		}
	}
}

// Test prompt with work in progress
func TestPromptWithProgressTracking(t *testing.T) {
	workProgress := `✓ Created models
✓ Added database migrations
• Pending: API endpoints
• Pending: Authentication`

	prompt := orchestrator.BuildResumePrompt(
		"Implement API",
		workProgress,
		"Next steps?",
		[]string{"Continue", "Review"},
		"Continue",
		"",
	)

	if !containsString(prompt, workProgress) {
		t.Errorf("prompt should include progress tracking")
	}
}

// Test prompt with user notes
func TestPromptWithUserNotes(t *testing.T) {
	userNotes := "Important: This needs to be backwards compatible and include comprehensive logging"

	prompt := orchestrator.BuildResumePrompt(
		"API Upgrade",
		"",
		"Which approach?",
		[]string{"Conservative", "Progressive"},
		"Conservative",
		userNotes,
	)

	if !containsString(prompt, userNotes) {
		t.Errorf("prompt should include user notes")
	}
}

// Test tasks of different statuses
func TestTaskStatusHandling(t *testing.T) {
	statuses := []types.Status{
		types.Pending,
		types.InProgress,
		types.NeedsReview,
		types.Completed,
	}

	for _, status := range statuses {
		task := types.Task{
			ID:     "task-1",
			Name:   "Test",
			Status: status,
		}

		statusStr := types.StatusString(task)
		if statusStr == "Unknown" {
			t.Errorf("status %d should have valid string representation", status)
		}
	}
}

// Test branch name length limits
func TestBranchNameLengthLimits(t *testing.T) {
	// Very long task name
	longTask := "This is an extremely long task description that contains way too many words and should be properly truncated to fit git branch naming conventions without being excessively long"

	branchName, _ := orchestrator.GenerateBranchName(longTask)

	// Should have reasonable length
	if len(branchName) > 100 {
		t.Errorf("branch name too long: %d characters", len(branchName))
	}

	// Should still be valid
	if branchName == "" {
		t.Errorf("branch name should not be empty after truncation")
	}

	if !hasPrefix(branchName, "ludwig/") {
		t.Errorf("truncated branch name should still have prefix")
	}
}

// Helper functions
func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func hasPrefix(str, prefix string) bool {
	return len(str) >= len(prefix) && str[:len(prefix)] == prefix
}
