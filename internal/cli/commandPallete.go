package cli
import (
	"ludwig/internal/utils"
	"fmt"
	"ludwig/internal/types"
	"ludwig/internal/storage"
	"github.com/google/uuid"
	"strings"
	"ludwig/internal/orchestrator"
	"strconv"
)

func PalleteCommands(taskStore *storage.FileTaskStorage) []utils.Command {
	return []utils.Command {
		{
			Text: "add",
			Action: func(text string) {
				parts := strings.Fields(text)
				if !checkArgumentsCountMin(2, parts, true) { return }

				// skip the first part which is the command itself
				newTask := &types.Task{
					Name: strings.Join(parts[1:], " "),
					Status: Pending,
					ID: uuid.New().String(),
				}

				if err := taskStore.AddTask(newTask); err != nil {
					fmt.Printf("Error adding new task: %v\n", err)
					return
				}
				// Refresh the kanban display
				GetTasksAndDisplayKanban(taskStore)
			},
			Description: "add <task description> - Add a new task. Tasks can be multiple words. No quotation marks needed.",
		},
		{
			Text: "start",
			Action: func(text string) {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) { return }
				utils.Println("Starting AI Orchestrator...")
				orchestrator.Start()
				utils.RequestAction(PalleteCommands(taskStore))
			},
			Description: "start - Start the AI Orchestrator",
		},
		{
			Text: "stop",
			Action: func(text string) {
				parts := strings.Fields(text)
				if len(parts) > 1 {
					utils.Println("Usage: stop method takes no arguments")
					return
				}
				utils.Println("Stopping AI Orchestrator...")
				orchestrator.Stop()
				utils.RequestAction(PalleteCommands(taskStore))
			},
			Description: "stop - Stop the AI Orchestrator",
		},
		{
			Text: "clear",
			Description: "clear - Clear the command line so that only the kanban board is visible",
			Action: func(text string) {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) { return }

				GetTasksAndDisplayKanban(taskStore)
			},
		},
		{
			Text: "exit",
			Description: "exit - Exit the CLI",
			Action: func(text string) {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) { return }

				utils.Println("Exiting CLI...")
				return
			},
		},
	}
}

func checkArgumentsCount(expected int, parts []string) bool {
	return checkArgumentsCountMin(expected, parts, false)
}

func checkArgumentsCountMin(expected int, parts []string, canHaveMore bool) bool {
	if canHaveMore && len(parts) >= expected { return true }
	if len(parts) != expected {
		utils.Println("Usage: " + parts[0] + " takes " + strconv.Itoa(expected) + " arguments")
		return false
	}
	return true
}
