package types_test

import (
	"testing"
	"time"

	"ludwig/internal/types"
)

// Test PrintTasks function
func TestPrintTasks(t *testing.T) {
	tasks := []types.Task{
		{
			ID:     "1",
			Name:   "Task 1",
			Status: types.Pending,
		},
		{
			ID:     "2",
			Name:   "Task 2",
			Status: types.InProgress,
		},
		{
			ID:     "3",
			Name:   "Task 3",
			Status: types.Completed,
		},
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTasks panicked: %v", r)
		}
	}()

	types.PrintTasks(tasks)
}

// Test PrintTasks with empty list
func TestPrintTasksEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTasks should handle empty list")
		}
	}()

	types.PrintTasks([]types.Task{})
}

// Test StatusString with all statuses
func TestStatusStringAllStatuses(t *testing.T) {
	testCases := []struct {
		status   types.Status
		expected string
	}{
		{types.Pending, "Pending"},
		{types.InProgress, "In Progress"},
		{types.NeedsReview, "In Review"},
		{types.Completed, "Completed"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			task := types.Task{Status: tc.status}
			result := types.StatusString(task)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// Test Task with all fields populated
func TestTaskFullyPopulated(t *testing.T) {
	now := time.Now()
	task := types.Task{
		ID:         "full-task",
		Name:       "Full Task",
		Status:     types.InProgress,
		BranchName: "feature/test",
		WorkInProgress: "Some progress",
		Review: &types.ReviewRequest{
			Question:  "Continue?",
			Context:   "Need decision",
			CreatedAt: now,
			Options: []types.ReviewOption{
				{ID: "y", Label: "Yes"},
				{ID: "n", Label: "No"},
			},
		},
		ReviewResponse: &types.ReviewResponse{
			ChosenOptionID: "y",
			ChosenLabel:    "Yes",
			UserNotes:      "Good approach",
			RespondedAt:    now,
		},
		ResponseFile: "responses/file.md",
	}

	if task.ID != "full-task" {
		t.Errorf("ID not set")
	}
	if task.Name != "Full Task" {
		t.Errorf("Name not set")
	}
	if task.BranchName != "feature/test" {
		t.Errorf("BranchName not set")
	}
	if task.Review == nil {
		t.Errorf("Review not set")
	}
	if task.ReviewResponse == nil {
		t.Errorf("ReviewResponse not set")
	}
	if task.ResponseFile != "responses/file.md" {
		t.Errorf("ResponseFile not set")
	}
}

// Test ReviewOption structure
func TestReviewOption(t *testing.T) {
	opt := types.ReviewOption{
		ID:    "opt-1",
		Label: "Option Label",
	}

	if opt.ID != "opt-1" {
		t.Errorf("ID not set correctly")
	}
	if opt.Label != "Option Label" {
		t.Errorf("Label not set correctly")
	}
}

// Test ReviewRequest with empty options
func TestReviewRequestEmptyOptions(t *testing.T) {
	req := types.ReviewRequest{
		Question: "Question",
		Options:  []types.ReviewOption{},
	}

	if len(req.Options) != 0 {
		t.Errorf("expected empty options")
	}
}

// Test ReviewRequest with many options
func TestReviewRequestManyOptions(t *testing.T) {
	options := make([]types.ReviewOption, 10)
	for i := 0; i < 10; i++ {
		options[i] = types.ReviewOption{
			ID:    string(rune(i)),
			Label: "Option",
		}
	}

	req := types.ReviewRequest{
		Question: "Choose?",
		Options:  options,
	}

	if len(req.Options) != 10 {
		t.Errorf("expected 10 options, got %d", len(req.Options))
	}
}

// Test ReviewResponse timestamps
func TestReviewResponseTimestamps(t *testing.T) {
	now := time.Now()
	resp := types.ReviewResponse{
		ChosenOptionID: "a",
		ChosenLabel:    "Option A",
		UserNotes:      "Notes",
		RespondedAt:    now,
	}

	if resp.RespondedAt != now {
		t.Errorf("timestamp not set correctly")
	}

	if resp.RespondedAt.IsZero() {
		t.Errorf("timestamp should not be zero")
	}
}

// Test ReviewRequest timestamps
func TestReviewRequestTimestamps(t *testing.T) {
	now := time.Now()
	req := types.ReviewRequest{
		Question:  "Q",
		CreatedAt: now,
	}

	if req.CreatedAt != now {
		t.Errorf("timestamp not set correctly")
	}

	if req.CreatedAt.IsZero() {
		t.Errorf("timestamp should not be zero")
	}
}

// Test Task status transitions
func TestTaskStatusTransitions(t *testing.T) {
	task := types.Task{
		ID:     "transit",
		Name:   "Transition test",
		Status: types.Pending,
	}

	transitions := []types.Status{
		types.InProgress,
		types.NeedsReview,
		types.Completed,
	}

	for i, newStatus := range transitions {
		task.Status = newStatus
		result := types.StatusString(task)
		if result == "Unknown" {
			t.Errorf("transition %d: status became unknown", i)
		}
	}
}

// Test ExampleTasks independence
func TestExampleTasksIndependence(t *testing.T) {
	tasks1 := types.ExampleTasks()
	tasks2 := types.ExampleTasks()

	// Modify first list
	tasks1[0].Name = "Modified"

	// Second list should be unchanged
	if tasks2[0].Name == "Modified" {
		t.Errorf("modifying one ExampleTasks list affected another")
	}
}

// Test task with nil review
func TestTaskWithNilReview(t *testing.T) {
	task := types.Task{
		ID:     "nil-review",
		Name:   "No review",
		Status: types.Completed,
		Review: nil,
	}

	if task.Review != nil {
		t.Errorf("review should be nil")
	}
}

// Test task with nil review response
func TestTaskWithNilReviewResponse(t *testing.T) {
	task := types.Task{
		ID:             "nil-response",
		Name:           "No response",
		Status:         types.Pending,
		ReviewResponse: nil,
	}

	if task.ReviewResponse != nil {
		t.Errorf("review response should be nil")
	}
}

// Test ReviewRequest context
func TestReviewRequestContext(t *testing.T) {
	req := types.ReviewRequest{
		Question: "Question?",
		Context:  "This is the context for the question",
		Options: []types.ReviewOption{
			{ID: "a", Label: "Option A"},
		},
	}

	if req.Context != "This is the context for the question" {
		t.Errorf("context not set correctly")
	}
}

// Test ReviewResponse user notes
func TestReviewResponseUserNotes(t *testing.T) {
	notes := "User provided these detailed notes about their choice"
	resp := types.ReviewResponse{
		ChosenOptionID: "a",
		ChosenLabel:    "Option A",
		UserNotes:      notes,
	}

	if resp.UserNotes != notes {
		t.Errorf("user notes not set correctly")
	}
}

// Test ReviewResponse with empty notes
func TestReviewResponseEmptyNotes(t *testing.T) {
	resp := types.ReviewResponse{
		ChosenOptionID: "a",
		ChosenLabel:    "Option A",
		UserNotes:      "",
	}

	if resp.UserNotes != "" {
		t.Errorf("expected empty notes")
	}
}

// Test WorkInProgress field
func TestTaskWorkInProgress(t *testing.T) {
	wip := "✓ Created auth module\n✓ Added JWT support\n• Pending: Rate limiting"
	
	task := types.Task{
		ID:             "wip-task",
		Name:           "Implementation",
		Status:         types.NeedsReview,
		WorkInProgress: wip,
	}

	if task.WorkInProgress != wip {
		t.Errorf("work in progress not set correctly")
	}
}

// Test BranchName field
func TestTaskBranchName(t *testing.T) {
	branchName := "feature/user-auth-system"
	
	task := types.Task{
		ID:         "branch-task",
		Name:       "Add auth",
		Status:     types.InProgress,
		BranchName: branchName,
	}

	if task.BranchName != branchName {
		t.Errorf("branch name not set correctly")
	}
}

// Test ResponseFile field
func TestTaskResponseFile(t *testing.T) {
	responseFile := "responses/task-123-20240101-120000.md"
	
	task := types.Task{
		ID:           "resp-task",
		Name:         "Task with response",
		Status:       types.Completed,
		ResponseFile: responseFile,
	}

	if task.ResponseFile != responseFile {
		t.Errorf("response file not set correctly")
	}
}

// Test multiple ReviewOptions in sequence
func TestMultipleReviewOptions(t *testing.T) {
	options := []types.ReviewOption{
		{ID: "opt1", Label: "First option"},
		{ID: "opt2", Label: "Second option"},
		{ID: "opt3", Label: "Third option"},
	}

	for i, opt := range options {
		if opt.ID == "" {
			t.Errorf("option %d ID is empty", i)
		}
		if opt.Label == "" {
			t.Errorf("option %d Label is empty", i)
		}
	}
}
