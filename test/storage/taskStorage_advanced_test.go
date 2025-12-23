package storage_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"ludwig/internal/storage"
	"ludwig/internal/types"
)

// Test sequential writes (simulating task operations)
func TestTaskStorageSequentialWrites(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()
	numTasks := 5

	for i := 1; i <= numTasks; i++ {
		task := &types.Task{
			ID:     string(rune(i)),
			Name:   "Task " + string(rune(i)),
			Status: types.Pending,
		}
		if err := s.AddTask(task); err != nil {
			t.Errorf("failed to add task: %v", err)
		}
	}

	// Verify all tasks were added
	tasks, _ := s.ListTasks()
	if len(tasks) != numTasks {
		t.Errorf("expected %d tasks after sequential writes, got %d", numTasks, len(tasks))
	}
}

// Test concurrent reads
func TestTaskStorageConcurrentReads(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	// Add some test data first
	task := &types.Task{ID: "shared-task", Name: "Shared", Status: types.Pending}
	s.AddTask(task)

	var wg sync.WaitGroup
	numReaders := 10
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.GetTask("shared-task")
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != numReaders {
		t.Errorf("expected %d successful reads, got %d", numReaders, successCount)
	}
}

// Test update with all fields
func TestTaskStorageUpdateAllFields(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	original := &types.Task{
		ID:     "full-task",
		Name:   "Original Name",
		Status: types.Pending,
	}
	s.AddTask(original)

	updated := &types.Task{
		ID:             "full-task",
		Name:           "Updated Name",
		Status:         types.Completed,
		BranchName:     "feature/new-branch",
		WorkInProgress: "Some work done",
		ResponseFile:   "responses/task.md",
	}

	s.UpdateTask(updated)
	retrieved, _ := s.GetTask("full-task")

	if retrieved.Name != "Updated Name" {
		t.Errorf("name not updated")
	}
	if retrieved.Status != types.Completed {
		t.Errorf("status not updated")
	}
	if retrieved.BranchName != "feature/new-branch" {
		t.Errorf("branch name not updated")
	}
	if retrieved.WorkInProgress != "Some work done" {
		t.Errorf("work in progress not updated")
	}
	if retrieved.ResponseFile != "responses/task.md" {
		t.Errorf("response file not updated")
	}
}

// Test task with review request
func TestTaskStorageWithReviewRequest(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{
		ID:     "review-task",
		Name:   "Task needing review",
		Status: types.NeedsReview,
		Review: &types.ReviewRequest{
			Question: "Which approach?",
			Options: []types.ReviewOption{
				{ID: "a", Label: "Approach A"},
				{ID: "b", Label: "Approach B"},
			},
			Context: "Need decision",
		},
	}

	s.AddTask(task)
	retrieved, _ := s.GetTask("review-task")

	if retrieved.Review == nil {
		t.Errorf("review request not persisted")
	}
	if retrieved.Review.Question != "Which approach?" {
		t.Errorf("review question not correct")
	}
	if len(retrieved.Review.Options) != 2 {
		t.Errorf("review options not correct")
	}
}

// Test task with review response
func TestTaskStorageWithReviewResponse(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{
		ID:     "response-task",
		Name:   "Task with response",
		Status: types.Completed,
		ReviewResponse: &types.ReviewResponse{
			ChosenOptionID: "a",
			ChosenLabel:    "Approach A",
			UserNotes:      "User preferred this",
		},
	}

	s.AddTask(task)
	retrieved, _ := s.GetTask("response-task")

	if retrieved.ReviewResponse == nil {
		t.Errorf("review response not persisted")
	}
	if retrieved.ReviewResponse.ChosenLabel != "Approach A" {
		t.Errorf("chosen label not correct")
	}
}

// Test large number of tasks
func TestTaskStorageLargeDataSet(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()
	numTasks := 100

	// Add many tasks
	for i := 0; i < numTasks; i++ {
		task := &types.Task{
			ID:     string(rune(i)),
			Name:   "Task " + string(rune(i)),
			Status: types.Pending,
		}
		s.AddTask(task)
	}

	// Verify all were added
	tasks, _ := s.ListTasks()
	if len(tasks) != numTasks {
		t.Errorf("expected %d tasks, got %d", numTasks, len(tasks))
	}

	// Delete some and verify
	for i := 0; i < 10; i++ {
		s.DeleteTask(string(rune(i)))
	}

	tasks, _ = s.ListTasks()
	if len(tasks) != numTasks-10 {
		t.Errorf("expected %d tasks after deletion, got %d", numTasks-10, len(tasks))
	}
}

// Test that ListTasks returns all status types
func TestTaskStorageListTasksMultipleStatus(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	statuses := []types.Status{
		types.Pending,
		types.InProgress,
		types.NeedsReview,
		types.Completed,
	}

	for i, status := range statuses {
		task := &types.Task{
			ID:     string(rune(i)),
			Name:   "Task",
			Status: status,
		}
		s.AddTask(task)
	}

	tasks, _ := s.ListTasks()

	statusFound := make(map[types.Status]bool)
	for _, task := range tasks {
		statusFound[task.Status] = true
	}

	for _, status := range statuses {
		if !statusFound[status] {
			t.Errorf("status %d not found in stored tasks", status)
		}
	}
}

// Test task fields independence
func TestTaskFieldsIndependence(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task1 := &types.Task{
		ID:             "task1",
		Name:           "Task 1",
		Status:         types.Pending,
		BranchName:     "branch1",
		WorkInProgress: "work1",
	}

	task2 := &types.Task{
		ID:             "task2",
		Name:           "Task 2",
		Status:         types.InProgress,
		BranchName:     "branch2",
		WorkInProgress: "work2",
	}

	s.AddTask(task1)
	s.AddTask(task2)

	t1, _ := s.GetTask("task1")
	t2, _ := s.GetTask("task2")

	if t1.WorkInProgress == t2.WorkInProgress {
		t.Errorf("task fields should be independent")
	}
	if t1.BranchName == t2.BranchName {
		t.Errorf("branch names should be independent")
	}
}

// Test storage path creation
func TestTaskStoragePathHandling(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	cwd, _ := os.Getwd()
	expectedPath := filepath.Join(cwd, ".ludwig", "tasks.json")

	s, _ := storage.NewFileTaskStorage()
	task := &types.Task{
		ID:     "test",
		Name:   "Test",
		Status: types.Pending,
	}
	s.AddTask(task)

	// Verify file exists at expected location
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("tasks.json not created at expected path: %s", expectedPath)
	}
}

// Test storage with empty task list persistence
func TestTaskStorageEmptyListPersistence(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s1, _ := storage.NewFileTaskStorage()
	task := &types.Task{ID: "temp", Name: "Temp", Status: types.Pending}
	s1.AddTask(task)
	s1.DeleteTask("temp")

	// Create new instance
	s2, _ := storage.NewFileTaskStorage()
	tasks, _ := s2.ListTasks()

	if len(tasks) != 0 {
		t.Errorf("expected empty list after deleting all tasks, got %d", len(tasks))
	}
}

// Test GetTask error message
func TestTaskStorageGetTaskErrorMessage(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	_, err := s.GetTask("nonexistent-123")
	if err == nil {
		t.Errorf("expected error for nonexistent task")
	}
	if err.Error() == "" {
		t.Errorf("error message should not be empty")
	}
}

// Test multiple rapid operations
func TestTaskStorageRapidOperations(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	// Add, update, read in quick succession
	task := &types.Task{ID: "rapid", Name: "Original", Status: types.Pending}
	s.AddTask(task)

	task.Name = "Updated"
	task.Status = types.InProgress
	s.UpdateTask(task)

	retrieved, _ := s.GetTask("rapid")
	if retrieved.Name != "Updated" {
		t.Errorf("rapid update failed")
	}

	s.DeleteTask("rapid")
	_, err := s.GetTask("rapid")
	if err == nil {
		t.Errorf("rapid delete failed")
	}
}
