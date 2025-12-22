package cli

import (
	"fmt"
	"strings"
	types "ludwig/internal/types"
	"ludwig/internal/utils"
)

const (
	Pending types.Status = iota
	InProgress
	Completed
)

func seperateTaskByStatus(tasks []types.Task) map[types.Status][]types.Task {
	taskLists := map[types.Status][]types.Task{
		Pending:    {},
		InProgress: {},
		Completed:  {},
	}
	for _, task := range tasks {
		taskLists[task.Status] = append(taskLists[task.Status], task)
	}
	return taskLists
}

const TASK_NAME_LENGTH = 30
func printKanbanHeader() {
	fmt.Print(" " + strings.Repeat("╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ ", 3) + "\r\n")
	fmt.Print(KanbanTaskName("Pending") + KanbanTaskName("In Progress") + KanbanTaskName("Completed") + "\r\n")
	fmt.Print(" " + strings.Repeat("├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ ", 3) + "\r\n")
}

func printKanbanFooter() {
	fmt.Print(" " + strings.Repeat("╰" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╯ ", 3) + "\r\n")
}
func KanbanTaskName(name string) string {
	if (len(name) + 5 > TASK_NAME_LENGTH) {
		truncatedName := name[:TASK_NAME_LENGTH - 7] + "..."
		numSpaces := TASK_NAME_LENGTH - len(truncatedName) - 4
		return " │ " + truncatedName + strings.Repeat(" ", numSpaces) + "│"
	}

	numSpaces := TASK_NAME_LENGTH - len(name) - 4

	return " │ " + name + strings.Repeat(" ", numSpaces) + "│"
}

func DisplayKanban(tasks []types.Task) {
	utils.ClearScreen()
	printKanbanHeader()
	taskLists := seperateTaskByStatus(tasks)

	listLengths := []int{
		len(taskLists[Pending]),
		len(taskLists[InProgress]),
		len(taskLists[Completed]),
	}
	maxListLength := 0
	for _, length := range listLengths {
		if length > maxListLength {
			maxListLength = length
		}
	}

	for i := 0; i < maxListLength; i++ {
		var line strings.Builder
		for status := Pending; status <= Completed; status++ {
			if i >= len(taskLists[status]) {
				line.WriteString(KanbanTaskName(""))
				continue;
			}
			line.WriteString(KanbanTaskName(taskLists[status][i].Name))
		}
		fmt.Print(line.String() + " \r\n")
	}
	printKanbanFooter()
}

