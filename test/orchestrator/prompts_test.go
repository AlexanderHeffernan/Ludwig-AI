package orchestrator_test

import (
	"errors"
	"strings"
	"testing"

	"ludwig/internal/orchestrator"
)

func TestBuildTaskPrompt(t *testing.T) {
	taskName := "Create a new user API endpoint"
	prompt := orchestrator.BuildTaskPrompt(taskName)

	if !strings.Contains(prompt, taskName) {
		t.Errorf("expected prompt to contain task name, got: %s", prompt)
	}

	if !strings.Contains(prompt, "Task:") {
		t.Errorf("expected prompt to contain 'Task:' label")
	}

	if !strings.Contains(prompt, "GIT WORKFLOW") {
		t.Errorf("expected prompt to contain system instructions")
	}
}

func TestBuildResumePrompt(t *testing.T) {
	taskName := "Create database schema"
	workInProgress := "✓ Created users table with id, name, email columns"
	question := "Should we use timestamps for tracking user creation?"
	options := []string{"Yes, add created_at and updated_at", "No, skip timestamps for now"}
	chosenLabel := "Yes, add created_at and updated_at"
	userNotes := "We need audit trails"

	prompt := orchestrator.BuildResumePrompt(taskName, workInProgress, question, options, chosenLabel, userNotes)

	if !strings.Contains(prompt, taskName) {
		t.Errorf("expected prompt to contain original task")
	}

	if !strings.Contains(prompt, workInProgress) {
		t.Errorf("expected prompt to contain work in progress")
	}

	if !strings.Contains(prompt, question) {
		t.Errorf("expected prompt to contain the question")
	}

	if !strings.Contains(prompt, chosenLabel) {
		t.Errorf("expected prompt to contain chosen label")
	}

	if !strings.Contains(prompt, userNotes) {
		t.Errorf("expected prompt to contain user notes")
	}

	for _, opt := range options {
		if !strings.Contains(prompt, opt) {
			t.Errorf("expected prompt to contain option: %s", opt)
		}
	}
}

func TestBuildResumePromptWithoutNotes(t *testing.T) {
	taskName := "Test"
	workInProgress := "Some work"
	question := "Question?"
	options := []string{"Option 1", "Option 2"}
	chosenLabel := "Option 1"
	userNotes := ""

	prompt := orchestrator.BuildResumePrompt(taskName, workInProgress, question, options, chosenLabel, userNotes)

	if !strings.Contains(prompt, taskName) {
		t.Errorf("expected prompt to contain task name")
	}

	// User notes should not cause errors even if empty
	if err := validatePrompt(prompt); err != nil {
		t.Errorf("prompt validation failed: %v", err)
	}
}

func TestBuildResumePromptWithoutWorkInProgress(t *testing.T) {
	taskName := "Test"
	workInProgress := ""
	question := "Question?"
	options := []string{"Option 1"}
	chosenLabel := "Option 1"
	userNotes := ""

	prompt := orchestrator.BuildResumePrompt(taskName, workInProgress, question, options, chosenLabel, userNotes)

	if !strings.Contains(prompt, taskName) {
		t.Errorf("expected prompt to contain task name")
	}

	if err := validatePrompt(prompt); err != nil {
		t.Errorf("prompt validation failed: %v", err)
	}
}

func TestBuildResumePromptMultipleOptions(t *testing.T) {
	options := []string{
		"Use PostgreSQL for primary database",
		"Use MongoDB for primary database",
		"Use DynamoDB for primary database",
	}
	prompt := orchestrator.BuildResumePrompt("DB choice", "", "Which database?", options, "Use PostgreSQL", "")

	for _, opt := range options {
		if !strings.Contains(prompt, opt) {
			t.Errorf("expected prompt to contain option: %s", opt)
		}
	}
}

// Helper function to validate prompt structure
func validatePrompt(prompt string) error {
	if prompt == "" {
		return errors.New("prompt is empty")
	}
	return nil
}
