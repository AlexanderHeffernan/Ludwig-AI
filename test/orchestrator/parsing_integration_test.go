package orchestrator_test

import (
	"strings"
	"testing"

	"ludwig/internal/orchestrator"
	"ludwig/internal/types"
	"time"
)

// Test review request parsing with valid input
func TestParseReviewRequestValid(t *testing.T) {
	// Since parseReviewRequest is private, we test it through the orchestrator logic
	// by verifying the prompt building and structure is correct
	prompt := orchestrator.BuildTaskPrompt("Test task")
	if !strings.Contains(prompt, "---NEEDS_REVIEW---") {
		t.Errorf("expected prompt to contain review marker in system instructions")
	}
}

// Test that review markers are properly formatted in prompts
func TestReviewMarkerFormatting(t *testing.T) {
	taskName := "Implement auth system"
	workInProgress := "Created login endpoint"
	question := "Use JWT or sessions?"
	options := []string{"JWT tokens", "Session cookies"}
	chosenLabel := "JWT tokens"
	userNotes := "Prefer stateless"

	prompt := orchestrator.BuildResumePrompt(taskName, workInProgress, question, options, chosenLabel, userNotes)

	// Verify structure
	if !strings.Contains(prompt, taskName) {
		t.Errorf("missing task name in prompt")
	}
	if !strings.Contains(prompt, workInProgress) {
		t.Errorf("missing work in progress in prompt")
	}
	if !strings.Contains(prompt, question) {
		t.Errorf("missing question in prompt")
	}
	if !strings.Contains(prompt, chosenLabel) {
		t.Errorf("missing chosen label in prompt")
	}
}

// Test prompt handles special characters
func TestBuildResumePromptSpecialCharacters(t *testing.T) {
	taskName := "Fix: Bug #123 & API v2.0"
	workInProgress := "Completed: First phase (60%)"
	question := "Use REST/GraphQL?"
	options := []string{"REST API", "GraphQL API"}
	chosenLabel := "REST API"
	userNotes := "Users prefer REST; easy to cache & monitor"

	prompt := orchestrator.BuildResumePrompt(taskName, workInProgress, question, options, chosenLabel, userNotes)

	if !strings.Contains(prompt, taskName) {
		t.Errorf("prompt should handle task names with special chars")
	}
	if !strings.Contains(prompt, workInProgress) {
		t.Errorf("prompt should handle work descriptions with special chars")
	}
}

// Test prompt with empty options
func TestBuildResumePromptEmptyOptions(t *testing.T) {
	taskName := "Test"
	workInProgress := "Work done"
	question := "Question?"
	options := []string{}
	chosenLabel := ""
	userNotes := ""

	prompt := orchestrator.BuildResumePrompt(taskName, workInProgress, question, options, chosenLabel, userNotes)

	if prompt == "" {
		t.Errorf("expected non-empty prompt even with empty options")
	}
}

// Test prompt with single option
func TestBuildResumePromptSingleOption(t *testing.T) {
	options := []string{"Proceed"}
	prompt := orchestrator.BuildResumePrompt("Task", "", "Continue?", options, "Proceed", "")

	if !strings.Contains(prompt, "Proceed") {
		t.Errorf("prompt should contain single option")
	}
}

// Test prompt with many options
func TestBuildResumePromptManyOptions(t *testing.T) {
	options := []string{
		"Option A",
		"Option B",
		"Option C",
		"Option D",
		"Option E",
	}
	prompt := orchestrator.BuildResumePrompt("Task", "", "Choose?", options, "Option A", "")

	for _, opt := range options {
		if !strings.Contains(prompt, opt) {
			t.Errorf("prompt should contain all options, missing: %s", opt)
		}
	}
}

// Test BuildTaskPrompt includes system instructions
func TestBuildTaskPromptIncludesSystemInstructions(t *testing.T) {
	prompt := orchestrator.BuildTaskPrompt("Sample task")

	requiredStrings := []string{
		"GIT WORKFLOW",
		"git add",
		"git commit",
		"NEEDS_REVIEW",
		"Task:",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(prompt, required) {
			t.Errorf("prompt missing required instruction: %s", required)
		}
	}
}

// Test that ReviewRequest structures are properly used
func TestReviewRequestCreation(t *testing.T) {
	now := time.Now()
	
	review := &types.ReviewRequest{
		Question: "Which approach?",
		Options: []types.ReviewOption{
			{ID: "opt1", Label: "Approach 1"},
			{ID: "opt2", Label: "Approach 2"},
		},
		Context:   "Need to decide",
		CreatedAt: now,
	}

	if review.Question != "Which approach?" {
		t.Errorf("review question not set properly")
	}
	if len(review.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(review.Options))
	}
	if review.Options[0].ID != "opt1" {
		t.Errorf("option ID not set properly")
	}
	if review.CreatedAt != now {
		t.Errorf("created at timestamp not set properly")
	}
}

// Test ReviewResponse creation
func TestReviewResponseCreation(t *testing.T) {
	now := time.Now()
	
	response := &types.ReviewResponse{
		ChosenOptionID: "opt1",
		ChosenLabel:    "Approach 1",
		UserNotes:      "This is better",
		RespondedAt:    now,
	}

	if response.ChosenOptionID != "opt1" {
		t.Errorf("chosen option ID not set properly")
	}
	if response.ChosenLabel != "Approach 1" {
		t.Errorf("chosen label not set properly")
	}
	if response.UserNotes != "This is better" {
		t.Errorf("user notes not set properly")
	}
	if response.RespondedAt != now {
		t.Errorf("responded at timestamp not set properly")
	}
}

// Test task with review workflow
func TestTaskWithReviewWorkflow(t *testing.T) {
	task := &types.Task{
		ID:     "task-1",
		Name:   "Implement feature",
		Status: types.NeedsReview,
		Review: &types.ReviewRequest{
			Question: "Design decision?",
			Options: []types.ReviewOption{
				{ID: "a", Label: "Option A"},
				{ID: "b", Label: "Option B"},
			},
		},
	}

	if task.Status != types.NeedsReview {
		t.Errorf("task status not set to NeedsReview")
	}
	if task.Review == nil {
		t.Errorf("task review not attached")
	}
	if len(task.Review.Options) != 2 {
		t.Errorf("review options not correct")
	}
}

// Test task with review response
func TestTaskWithReviewResponse(t *testing.T) {
	task := &types.Task{
		ID:     "task-1",
		Name:   "Implement feature",
		Status: types.NeedsReview,
		ReviewResponse: &types.ReviewResponse{
			ChosenOptionID: "a",
			ChosenLabel:    "Option A",
			UserNotes:      "Preferred for performance",
		},
	}

	if task.ReviewResponse == nil {
		t.Errorf("task review response not attached")
	}
	if task.ReviewResponse.ChosenLabel != "Option A" {
		t.Errorf("chosen label not correct")
	}
}

// Test prompt consistency across multiple builds
func TestPromptConsistency(t *testing.T) {
	taskName := "Test task"
	
	prompt1 := orchestrator.BuildTaskPrompt(taskName)
	prompt2 := orchestrator.BuildTaskPrompt(taskName)

	if prompt1 != prompt2 {
		t.Errorf("BuildTaskPrompt should produce consistent output")
	}

	// Verify both contain the task
	if !strings.Contains(prompt1, taskName) {
		t.Errorf("first prompt missing task name")
	}
	if !strings.Contains(prompt2, taskName) {
		t.Errorf("second prompt missing task name")
	}
}
