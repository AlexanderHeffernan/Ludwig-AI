package cli

import (
	"fmt"
	types "ludwig/internal/types"
	"ludwig/internal/utils"
)

func Execute() {
	fmt.Println("Starting AI Orchestrator CLI...")
	// Call orchestrator and mcp as needed
	var tasks = types.ExampleTasks()
	DisplayKanban(tasks)

	utils.OnKeyPress([]utils.KeyAction{
		{
			Key: 'a',
			Action: func() {
				fmt.Println("Adding a new task...")
			},
			Description: "New Task",
		},
	})
}
