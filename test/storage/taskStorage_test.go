package storage_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ludwig/internal/storage"
	"ludwig/internal/types"
)

func cleanupTestStorage(t *testing.T) {
	cwd, _ := os.Getwd()
	ludwigDir := filepath.Join(cwd, ".ludwig")
	os.RemoveAll(ludwigDir)
}

func setupTestStorage(t *testing.T) {
	cleanupTestStorage(t)
}

func TestNewFileTaskStorageCreatesEmptyStorage(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, err := storage.NewFileTaskStorage()
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if s == nil {
		t.Errorf("expected storage, got nil")
	}

	tasks, err := s.ListTasks()
	if err != nil {
		t.Errorf("failed to list tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks initially, got %d", len(tasks))
	}
}

func TestAddAndGetTask(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{
		ID:     "task-1",
		Name:   "Test Task",
		Status: types.Pending,
	}

	if err := s.AddTask(task); err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	retrieved, err := s.GetTask("task-1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if retrieved.ID != "task-1" {
		t.Errorf("expected ID task-1, got %s", retrieved.ID)
	}
	if retrieved.Name != "Test Task" {
		t.Errorf("expected name 'Test Task', got %s", retrieved.Name)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	_, err := s.GetTask("nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent task, got nil")
	}
}

func TestListTasks(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	tasks := []*types.Task{
		{ID: "1", Name: "Task 1", Status: types.Pending},
		{ID: "2", Name: "Task 2", Status: types.InProgress},
		{ID: "3", Name: "Task 3", Status: types.Completed},
	}

	for _, task := range tasks {
		if err := s.AddTask(task); err != nil {
			t.Fatalf("failed to add task: %v", err)
		}
	}

	retrieved, err := s.ListTasks()
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(retrieved))
	}
}

func TestUpdateTask(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{
		ID:     "task-1",
		Name:   "Original Name",
		Status: types.Pending,
	}

	s.AddTask(task)

	updated := &types.Task{
		ID:     "task-1",
		Name:   "Updated Name",
		Status: types.InProgress,
	}

	if err := s.UpdateTask(updated); err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	retrieved, _ := s.GetTask("task-1")
	if retrieved.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Status != types.InProgress {
		t.Errorf("expected status InProgress, got %v", retrieved.Status)
	}
}

func TestUpdateTaskNotFound(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{ID: "nonexistent", Name: "Test"}

	err := s.UpdateTask(task)
	if err == nil {
		t.Errorf("expected error updating nonexistent task, got nil")
	}
}

func TestDeleteTask(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{
		ID:     "task-1",
		Name:   "Test Task",
		Status: types.Pending,
	}

	s.AddTask(task)

	if err := s.DeleteTask("task-1"); err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	_, err := s.GetTask("task-1")
	if err == nil {
		t.Errorf("expected error getting deleted task, got nil")
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	err := s.DeleteTask("nonexistent")
	if err == nil {
		t.Errorf("expected error deleting nonexistent task, got nil")
	}
}

func TestTaskPersistence(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	// Create first storage instance and add task
	s1, _ := storage.NewFileTaskStorage()
	task := &types.Task{
		ID:     "persist-test",
		Name:   "Persistence Test",
		Status: types.InProgress,
	}
	s1.AddTask(task)

	// Create new storage instance and verify task is loaded
	s2, _ := storage.NewFileTaskStorage()
	retrieved, err := s2.GetTask("persist-test")
	if err != nil {
		t.Fatalf("failed to retrieve persisted task: %v", err)
	}

	if retrieved.Name != "Persistence Test" {
		t.Errorf("expected persisted name, got %s", retrieved.Name)
	}
}

func TestTaskJSONFormat(t *testing.T) {
	setupTestStorage(t)
	defer cleanupTestStorage(t)

	s, _ := storage.NewFileTaskStorage()

	task := &types.Task{
		ID:     "format-test",
		Name:   "Format Test",
		Status: types.Completed,
	}

	s.AddTask(task)

	// Read the JSON file directly
	cwd, _ := os.Getwd()
	taskFile := filepath.Join(cwd, ".ludwig", "tasks.json")
	data, err := os.ReadFile(taskFile)
	if err != nil {
		t.Fatalf("failed to read task file: %v", err)
	}

	var tasks map[string]*types.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, ok := tasks["format-test"]; !ok {
		t.Errorf("task not found in JSON file")
	}
}
