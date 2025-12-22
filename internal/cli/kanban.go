package cli

import (
	"fmt"
	"strings"
	types "ludwig/internal/types"
	"ludwig/internal/utils"
)

func seperateTaskByStatus(tasks []types.Task) map[types.Status][]types.Task {
	taskLists := map[types.Status][]types.Task{
		types.Pending:     {},
		types.InProgress:  {},
		types.NeedsReview: {},
		types.Completed:   {},
	}
	for _, task := range tasks {
		taskLists[task.Status] = append(taskLists[task.Status], task)
	}
	return taskLists
}

const TASK_NAME_LENGTH = 30
func printKanbanHeader() {
	//fmt.Print(" " + strings.Repeat("╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ ", 4) + "\r\n")
	//fmt.Print(KanbanTaskName("Pending") + KanbanTaskName("In Progress") + KanbanTaskName("In Review") + KanbanTaskName("Completed") + "\r\n")
	//fmt.Print(" " + strings.Repeat("├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ ", 4) + "\r\n")
	fmt.Print(genKanbanHeader())
}

func genKanbanHeader() string {
	var header strings.Builder
	header.WriteString(" " + strings.Repeat("╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ ", 4) + "\n")
	header.WriteString(KanbanTaskName("Pending") + KanbanTaskName("In Progress") + KanbanTaskName("In Review") + KanbanTaskName("Completed") + "\n")
	header.WriteString(" " + strings.Repeat("├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ ", 4) + "\n")
	return header.String()
}

func printKanbanFooter() {
	fmt.Print(genKanbanFooter())
}

func genKanbanFooter() string {
	return " " + strings.Repeat("╰" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╯ ", 4) + "\n"
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
		len(taskLists[types.Pending]),
		len(taskLists[types.InProgress]),
		len(taskLists[types.NeedsReview]),
		len(taskLists[types.Completed]),
	}
	maxListLength := 0
	for _, length := range listLengths {
		if length > maxListLength {
			maxListLength = length
		}
	}

	for i := 0; i < maxListLength; i++ {
		var line strings.Builder
		for status := types.Pending; status <= types.Completed; status++ {
			if i >= len(taskLists[status]) {
				line.WriteString(KanbanTaskName(""))
				continue;
			}
			line.WriteString(KanbanTaskName(taskLists[status][i].Name))
		}
		fmt.Print(line.String() + " \n")
	}
	printKanbanFooter()
}

func RenderKanban(tasks []types.Task) string {
	var builder strings.Builder
	//printKanbanHeader()
	builder.WriteString(genKanbanHeader())
	taskLists := seperateTaskByStatus(tasks)

	listLengths := []int{
		len(taskLists[types.Pending]),
		len(taskLists[types.InProgress]),
		len(taskLists[types.NeedsReview]),
		len(taskLists[types.Completed]),
	}
	maxListLength := 0
	for _, length := range listLengths {
		if length > maxListLength {
			maxListLength = length
		}
	}

	for i := 0; i < maxListLength; i++ {
		var line strings.Builder
		for status := types.Pending; status <= types.Completed; status++ {
			if i >= len(taskLists[status]) {
				line.WriteString(KanbanTaskName(""))
				continue;
			}
			line.WriteString(KanbanTaskName(taskLists[status][i].Name))
		}
		builder.WriteString(line.String() + " \n")
	}
	builder.WriteString(genKanbanFooter())
	return builder.String()
}
