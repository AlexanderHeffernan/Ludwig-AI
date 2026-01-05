package model

import (
	"ludwig/internal/utils"
	"ludwig/internal/storage"
	"ludwig/internal/types/task"
	"ludwig/internal/orchestrator"

	"strings"
	"time"
	"strconv"
	"os"

	"github.com/google/uuid"
	"github.com/charmbracelet/bubbles/table"
)

func PalleteCommands(taskStore *storage.FileTaskStorage) []Command {
	actions := []Command {
		{
			Text: "add",
			Action: func(text string, m *Model) string {
				parts := strings.Fields(text)
				if !checkArgumentsCountMin(2, parts, true) {
					return "Usage: add <task description> - Add a new task. Tasks can be multiple words. No quotation marks needed."
				}

				// skip the first part which is the command itself
				newTask := &task.Task{
					Name: strings.Join(parts[1:], " "),
					Status: task.Pending,
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
			Text: "delete",
			Description: "delete <task ref> - Delete a task by it's ref, can be seen to the left of the task name on the kanban. Do not include the # symbol.",
			Action: func(text string, m *Model) string {
				parts := strings.Fields(text)
				if !checkArgumentsCount(2, parts) {
					return "Usage: delete <task ref> - Delete a task by it's ref, can be seen to the left of the task name on the kanban."
				}
				taskIndex, err := strconv.Atoi(parts[1])
				if err != nil {
					return "Invalid task ref. Must be a number."
				}
				
				tasksPointers, err := taskStore.ListTasks()
				if err != nil {
					return "Error retrieving tasks: " + err.Error()
				}

				tasks := utils.PointerSliceToValueSlice(tasksPointers)

				if taskIndex < 0 || taskIndex >= len(tasks) {
					return "Task ref out of range."
				}
				taskToDelete := tasks[taskIndex]
				if err := taskStore.DeleteTask(taskToDelete.ID); err != nil {
					return "Error deleting task: " + err.Error()
				}
				return "Deleted task: " + taskToDelete.Name
			},
		},
		{
			Text: "start",
			Action: func(text string, m *Model) string {
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
			Action: func(text string, m *Model) string {
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
			Action: func(text string, m *Model) string {
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
			Action: func(text string, m *Model) string {
				parts := strings.Fields(text)
				if !checkArgumentsCount(1, parts) {
					return "Usage: exit method takes no arguments"
				}

				//utils.Println("Exiting CLI...")
				os.Exit(0)
				return ""
			},
		},
		{
			Text: "view",
			Description: "view <task ref> - View the streamed output log of a task by it's ref. Do not include the # symbol.",
			Action: func(text string, m *Model) string {
				parts := strings.Fields(text)
				if !checkArgumentsCount(2, parts) {
					return "Usage: view command takes 1 argument: <task ref>"
				}

				taskIndex, err := strconv.Atoi(parts[1])
				if err != nil {
					return "Invalid task ref. Must be a number."
				}
				
				tasksPointers, err := taskStore.ListTasks()
				if err != nil {
					return "Error retrieving tasks: " + err.Error()
				}

				tasks := utils.PointerSliceToValueSlice(tasksPointers)

				if taskIndex < 0 || taskIndex >= len(tasks) {
					return "Task ref out of range."
				}
				taskToView := tasks[taskIndex]
				filePath := "./.ludwig/" + taskToView.ResponseFile

				m.viewingViewport = true
				m.taskViewport = *m.taskViewport.SetViewingTask(&taskToView, filePath)
				m.taskViewport.ViewportUpdateLoop()

				return ""
			},
		},
	}
	return append(actions, Command {
		Text: "help",
		Description: "help - Show this help message",
		Action: func(text string, m *Model) string {
			parts := strings.Fields(text)
			if !checkArgumentsCount(1, parts) {
				return "Usage: help method takes no arguments"
			}
			//utils.PrintHelp(actions)
			return PrintHelpTable(actions)
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

func PrintHelpTable(actions []Command) string {
	columns := []table.Column {
		{Title: "Command", Width: 20},
		{Title: "Description", Width: 200},
	}
	rows := genHelpTableRows(actions)
	t := tableOptions(columns, rows)

	return t.View()
}

func genHelpTableRows(actions []Command) []table.Row {
	var rows []table.Row
	for _, cmd := range actions {
		rows = append(rows, table.Row{cmd.Text, cmd.Description})
	}
	rows = append(rows, table.Row{"help", "Show this help message"})
	return rows
}

func tableOptions(columns []table.Column, rows []table.Row) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
	)
	s := table.DefaultStyles()
	s.Selected = s.Cell.Padding(0,0).Margin(0,0)
	t.SetStyles(s)
	return t
}
