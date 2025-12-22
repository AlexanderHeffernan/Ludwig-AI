package cli

import (
	"fmt"
	"ludwig/internal/utils"
	"ludwig/internal/storage"
)

func GetTasksAndDisplayKanban(taskStore *storage.FileTaskStorage) {
	tasks, err := taskStore.ListTasks()
	if err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
		return
	}
	DisplayKanban(utils.PointerSliceToValueSlice(tasks))
} 

func Execute() {
	fmt.Println("Starting AI Orchestrator CLI...")
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		fmt.Printf("Error initializing task storage: %v\n", err)
		return
	}

	for {
		GetTasksAndDisplayKanban(taskStore)
		result := utils.RequestAction(PalleteCommands(taskStore))
		if result == "exit" {
			break
		}
	}
}
