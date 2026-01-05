package types_test

import (
	"testing"
	"time"

	"ludwig/internal/types/task"
)

// Test PrintTasks function
func TestPrintTasks(t *testing.T) {
	tasks := []task.Task{
		{
			ID:     "1",
			Name:   "Task 1",
			Status: task.Pending,
		},
		{
			ID:     "2",
			Name:   "Task 2",
			Status: task.InProgress,
		},
		{
			ID:     "3",
			Name:   "Task 3",
			Status: task.Completed,
		},
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTasks panicked: %v", r)
		}
	}()

	task.PrintTasks(tasks)
}

// Test PrintTasks with empty list
func TestPrintTasksEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTasks should handle empty list")
		}
	}()

	task.PrintTasks([]task.Task{})
}

// Test StatusString with all statuses
func TestStatusStringAllStatuses(t *testing.T) {
	testCases := []struct {
		status   task.Status
		expected string
	}{
		{task.Pending, "Pending"},
		{task.InProgress, "In Progress"},
		{task.NeedsReview, "In Review"},
		{task.Completed, "Completed"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			testTask := task.Task{Status: tc.status}
			result := task.StatusString(testTask)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// Test Task with all fields populated
func TestTaskFullyPopulated(t *testing.T) {
	now := time.Now()
	testTask := task.Task{
		ID:         "full-task",
		Name:       "Full Task",
		Status:     task.InProgress,
		BranchName: "feature/test",
		WorkInProgress: "Some progress",
		Review: &task.ReviewRequest{
			Question:  "Continue?",
			Context:   "Need decision",
			CreatedAt: now,
			Options: []task.ReviewOption{
				{ID: "y", Label: "Yes"},
				{ID: "n", Label: "No"},
			},
		},
		ReviewResponse: &task.ReviewResponse{
			ChosenOptionID: "y",
			ChosenLabel:    "Yes",
			UserNotes:      "Good approach",
			RespondedAt:    now,
		},
		ResponseFile: "responses/file.md",
	}

	if testTask.ID != "full-task" {
		t.Errorf("ID not set")
	}
	if testTask.Name != "Full Task" {
		t.Errorf("Name not set")
	}
	if testTask.BranchName != "feature/test" {
		t.Errorf("BranchName not set")
	}
	if testTask.Review == nil {
		t.Errorf("Review not set")
	}
	if testTask.ReviewResponse == nil {
		t.Errorf("ReviewResponse not set")
	}
	if testTask.ResponseFile != "responses/file.md" {
		t.Errorf("ResponseFile not set")
	}
}

// Test ReviewOption structure
func TestReviewOption(t *testing.T) {
	opt := task.ReviewOption{
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
	req := task.ReviewRequest{
		Question: "Question",
		Options:  []task.ReviewOption{},
	}

	if len(req.Options) != 0 {
		t.Errorf("expected empty options")
	}
}

// Test ReviewRequest with many options
func TestReviewRequestManyOptions(t *testing.T) {
	options := make([]task.ReviewOption, 10)
	for i := 0; i < 10; i++ {
		options[i] = task.ReviewOption{
			ID:    string(rune(i)),
			Label: "Option",
		}
	}

	req := task.ReviewRequest{
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
	resp := task.ReviewResponse{
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
	req := task.ReviewRequest{
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
	testTask := task.Task{
		ID:     "transit",
		Name:   "Transition test",
		Status: task.Pending,
	}

	transitions := []task.Status{
		task.InProgress,
		task.NeedsReview,
		task.Completed,
	}

	for i, newStatus := range transitions {
		testTask.Status = newStatus
		result := task.StatusString(testTask)
		if result == "Unknown" {
			t.Errorf("transition %d: status became unknown", i)
		}
	}
}

// Test ExampleTasks independence
func TestExampleTasksIndependence(t *testing.T) {
	tasks1 := task.ExampleTasks()
	tasks2 := task.ExampleTasks()

	// Modify first list
	tasks1[0].Name = "Modified"

	// Second list should be unchanged
	if tasks2[0].Name == "Modified" {
		t.Errorf("modifying one ExampleTasks list affected another")
	}
}

// Test task with nil review
func TestTaskWithNilReview(t *testing.T) {
	testTask := task.Task{
		ID:     "nil-review",
		Name:   "No review",
		Status: task.Completed,
		Review: nil,
	}

	if testTask.Review != nil {
		t.Errorf("review should be nil")
	}
}

// Test task with nil review response
func TestTaskWithNilReviewResponse(t *testing.T) {
	testTask := task.Task{
		ID:             "nil-response",
		Name:           "No response",
		Status:         task.Pending,
		ReviewResponse: nil,
	}

	if testTask.ReviewResponse != nil {
		t.Errorf("review response should be nil")
	}
}

// Test ReviewRequest context
func TestReviewRequestContext(t *testing.T) {
	req := task.ReviewRequest{
		Question: "Question?",
		Context:  "This is the context for the question",
		Options: []task.ReviewOption{
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
	resp := task.ReviewResponse{
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
	resp := task.ReviewResponse{
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
	
	testTask := task.Task{
		ID:             "wip-task",
		Name:           "Implementation",
		Status:         task.NeedsReview,
		WorkInProgress: wip,
	}

	if testTask.WorkInProgress != wip {
		t.Errorf("work in progress not set correctly")
	}
}

// Test BranchName field
func TestTaskBranchName(t *testing.T) {
	branchName := "feature/user-auth-system"
	
	testTask := task.Task{
		ID:         "branch-task",
		Name:       "Add auth",
		Status:     task.InProgress,
		BranchName: branchName,
	}

	if testTask.BranchName != branchName {
		t.Errorf("branch name not set correctly")
	}
}

// Test ResponseFile field
func TestTaskResponseFile(t *testing.T) {
	responseFile := "responses/task-123-20240101-120000.md"
	
	testTask := task.Task{
		ID:           "resp-task",
		Name:         "Task with response",
		Status:       task.Completed,
		ResponseFile: responseFile,
	}

	if testTask.ResponseFile != responseFile {
		t.Errorf("response file not set correctly")
	}
}

// Test multiple ReviewOptions in sequence
func TestMultipleReviewOptions(t *testing.T) {
	options := []task.ReviewOption{
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
