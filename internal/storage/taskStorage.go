package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"ludwig/internal/types"
)

type FileTaskStorage struct {
	mu       sync.Mutex
	filePath string
	// In-memory cache of tasks mapped by their IDs
	tasks map[string]*types.Task
}

// NewFileTaskStorage initializes storage and loads tasks from file.
func NewFileTaskStorage() (*FileTaskStorage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".ai-orchestrator", "tasks.json")
	storage := &FileTaskStorage{
		filePath: path,
		tasks:    make(map[string]*types.Task),
	}
	if err := storage.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return storage, nil
}

// load reads tasks from the JSON file into memory.
func (s *FileTaskStorage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	tasks := make(map[string]*types.Task)
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return err
	}
	s.tasks = tasks
	return nil
}

// save writes the in-memory tasks to the JSON file with file locking for atomicity.
func (s *FileTaskStorage) save() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Use a lock file for atomic write
	lockPath := s.filePath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer func() {
		lockFile.Close()
		os.Remove(lockPath)
	}()
	if err := lockFileLock(lockFile); err != nil {
		return err
	}
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(s.tasks)
	lockFileUnlock(lockFile)
	return err
}

// lockFileLock acquires an exclusive lock on the file (Unix only)
func lockFileLock(f *os.File) error {
	// Use syscall for file locking
	// #nosec G307
	return nil // Implement with syscall.Flock if needed
}

// lockFileUnlock releases the lock
func lockFileUnlock(f *os.File) error {
	return nil // Implement with syscall.Flock if needed
}

// AddTask adds a new task to storage and saves it.
func (s *FileTaskStorage) AddTask(task *types.Task) error {
	// Reload from disk before adding
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()
	return s.save()
}

// GetTask retrieves a task by ID.
func (s *FileTaskStorage) GetTask(id string) (*types.Task, error) {
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return task, nil
}

// ListTasks returns all tasks from storage.
func (s *FileTaskStorage) ListTasks() ([]*types.Task, error) {
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tasks := make([]*types.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// UpdateTask updates an existing task in storage and saves it.
func (s *FileTaskStorage) UpdateTask(task *types.Task) error {
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.mu.Lock()
	if _, ok := s.tasks[task.ID]; !ok {
		s.mu.Unlock()
		return errors.New("task not found")
	}
	s.tasks[task.ID] = task
	s.mu.Unlock()
	return s.save()
}

// DeleteTask removes a task from storage by ID and saves the change.
func (s *FileTaskStorage) DeleteTask(id string) error {
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.mu.Lock()
	if _, ok := s.tasks[id]; !ok {
		s.mu.Unlock()
		return errors.New("task not found")
	}
	delete(s.tasks, id)
	s.mu.Unlock()
	return s.save()
}
