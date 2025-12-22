package orchestrator

import (
	"fmt"
	"log"

	"ludwig/internal/orchestrator/clients"
	"ludwig/internal/storage"
	"ludwig/internal/types"
)

func Start() {
	// Initialize task storage
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		log.Fatalf("Failed to initialize task storage: %v", err)
	}

	// Add dummy tasks
	// dummyTasks := []*types.Task{
	// 	{ID: "1", Name: "Create a file called hello.txt which says Hello, World!", Status: types.Pending},
	// 	{ID: "2", Name: "Create a file called goodbye.txt which says Goodbye, World!", Status: types.Pending},
	// }
	// for _, t := range dummyTasks {
	// 	_ = taskStore.AddTask(t) // Ignore error for demo; handle in real code
	// }

	// List tasks
	tasks, err := taskStore.ListTasks()
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	// Initialize Gemini client
	gemini := &clients.GeminiClient{}

	// Send each task's description to Gemini
	for _, t := range tasks {
		if t.Status != types.Pending {
			continue
		}
		response, err := gemini.SendPrompt(t.Name)
		if err != nil {
			fmt.Printf("Error sending task %s to Gemini: %v\n", t.ID, err)
			continue
		}
		fmt.Printf("Gemini response for task %s: %s\n", t.ID, response)

		// Update task status to Completed
		t.Status = types.Completed
		_ = taskStore.AddTask(t) // Ignore error for demo; handle in real code
	}
}
