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

	utils.RequestAction(PalleteCommands(taskStore))
} 

func Execute() {
	fmt.Println("Starting AI Orchestrator CLI...")
	// Call orchestrator and mcp as needed
	//var tasks = types.ExampleTasks()
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		fmt.Printf("Error initializing task storage: %v\n", err)
		return
	}
	GetTasksAndDisplayKanban(taskStore)

}
