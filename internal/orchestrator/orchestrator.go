package orchestrator

import (
	"log"
	"sync"
	"time"

	"ludwig/internal/orchestrator/clients"
	"ludwig/internal/storage"
	"ludwig/internal/types"
)

var (
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
)

// Start launches the orchestrator loop in a goroutine.
func Start() {
	mu.Lock()
	defer mu.Unlock()
	if running {
		return
	}
	running = true
	stopCh = make(chan struct{})
	wg.Add(1)
	go orchestratorLoop()
}

// Stop signals the orchestrator to stop and waits for it to finish.
func Stop() {
	mu.Lock()
	if !running {
		mu.Unlock()
		return
	}
	close(stopCh)
	mu.Unlock()
	wg.Wait()
	mu.Lock()
	running = false
	mu.Unlock()
}

// IsRunning returns true if the orchestrator is running.
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}

// orchestratorLoop processes tasks, polling for new ones.
func orchestratorLoop() {
	defer wg.Done()
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		log.Printf("Failed to initialize task storage: %v", err)
		return
	}
	gemini := &clients.GeminiClient{}

	for {
		select {
		case <-stopCh:
			return
		default:
			// Process one pending task at a time
			tasks, err := taskStore.ListTasks()
			if err != nil {
				log.Printf("Failed to list tasks: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			processed := false
			for _, t := range tasks {
				if t.Status == types.Pending {
					log.Printf("Starting task %s: %s", t.ID, t.Name)
					// Set status to InProgress before processing
					t.Status = types.InProgress
					if err := taskStore.UpdateTask(t); err != nil {
						log.Printf("Failed to set task %s to In Progress: %v", t.ID, err)
						continue
					}
					response, err := gemini.SendPrompt(t.Name)
					if err != nil {
						log.Printf("Error sending task %s to Gemini: %v", t.ID, err)
						// Optionally set back to Pending if failed
						t.Status = types.Pending
						_ = taskStore.UpdateTask(t)
						continue
					}
					log.Printf("Completed task %s: Gemini response: %s", t.ID, response)
					t.Status = types.Completed
					_ = taskStore.UpdateTask(t)
					processed = true
					break // Only process one task per loop
				}
			}
			if !processed {
				log.Printf("No pending tasks found. Waiting before polling again.")
				time.Sleep(2 * time.Second) // No pending tasks, wait before polling again
			}
		}
	}
}
