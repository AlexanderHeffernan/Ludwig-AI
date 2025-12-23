package cli
import (
	"ludwig/internal/utils"
	//"fmt"
	"ludwig/internal/types"
	"ludwig/internal/storage"
	"github.com/google/uuid"
	"strings"
	"ludwig/internal/orchestrator"
	"strconv"
	"os"
	"time"
)

func PalleteCommands(taskStore *storage.FileTaskStorage) []utils.Command {
	actions := []utils.Command {
		{
			Text: "add",
			Action: func(text string) string {
				parts := strings.Fields(text)
				if !checkArgumentsCountMin(2, parts, true) {
					return "Usage: add <task description> - Add a new task. Tasks can be multiple words. No quotation marks needed."
				}

				// skip the first part which is the command itself
				newTask := &types.Task{
					Name: strings.Join(parts[1:], " "),
					Status: types.Pending,
					ID: uuid.New().String(),
					CreatedAt: time.Now(),
				}

				if err := taskStore.AddTask(newTask); err != nil {
					//fmt.Printf("Error adding new task: %v\n", err)
					return "Error adding new task: " + err.Error()
				}
				return "Added new task: " + newTask.Name
			},
			Description: "add <task description> - Add a new task. Tasks can be multiple words. No quotation marks needed.",
		},
		{
			Text: "start",
			Action: func(text string) string {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) {
					return "Usage: start method takes no arguments"
				}
				orchestrator.Start()
				return "AI Orchestrator started."
			},
			Description: "start - Start the AI Orchestrator",
		},
		{
			Text: "stop",
			Action: func(text string) string {
				parts := strings.Fields(text)
				if len(parts) > 1 {
					//utils.Println("Usage: stop method takes no arguments")
					return "Usage: stop method takes no arguments"
				}
				//utils.Println("Stopping AI Orchestrator...")
				orchestrator.Stop()
				return "AI Orchestrator stopped."
			},
			Description: "stop - Stop the AI Orchestrator",
		},
		{
			Text: "clear",
			Description: "clear - Clear the command line so that only the kanban board is visible",
			Action: func(text string) string {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) {
					return  "Usage: clear method takes no arguments"
				}
				return ""
			},
		},
		{
			Text: "exit",
			Description: "exit - Exit the CLI",
			Action: func(text string) string {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) {
					return "Usage: exit method takes no arguments"
				}

				//utils.Println("Exiting CLI...")
				os.Exit(0)
				return ""
			},
		},
	}
	return append(actions, utils.Command {
		Text: "help",
		Description: "help - Show this help message",
		Action: func(text string) string {
			parts := strings.Fields(text)
			if !checkArgumentsCount(1, parts) {
				return "Usage: help method takes no arguments"
			}
			//utils.PrintHelp(actions)
			return utils.PrintHelp(actions)
		},
	})
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
