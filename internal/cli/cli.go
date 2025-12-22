package cli

import (
	"fmt"
	"ludwig/internal/types"
	"ludwig/internal/utils"
	"ludwig/internal/storage"
	"github.com/google/uuid"
)

func GetTasksAndDisplayKanban(taskStore *storage.FileTaskStorage) {
	tasks, err := taskStore.ListTasks()
	if err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
		return
	}
	DisplayKanban(utils.PointerSliceToValueSlice(tasks))

	utils.OnKeyPress([]utils.KeyAction{
		{
			Key: 'a',
			Action: func() {
				description := utils.RequestInput("Enter task description")

				newTask := &types.Task{
					Name: description,
					Status: Pending,
					ID: uuid.New().String(),
				}
				
				if err := taskStore.AddTask(newTask); err != nil {
					fmt.Printf("Error adding new task: %v\n", err)
					return
				}
				//fmt.Printf("Added new task with ID: %s\n", newTask.ID)
				GetTasksAndDisplayKanban(taskStore)
			},
			Description: "New Task",
		},
	})
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
