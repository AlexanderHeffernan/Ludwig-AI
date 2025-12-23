package cli

import (
	"fmt"
	"strings"
	types "ludwig/internal/types"
	"ludwig/internal/utils"
)

var borderColors map[types.Status]string = map[types.Status]string {
	types.Pending:     "34", // Blue
	types.InProgress:  "33", // Yellow
	types.NeedsReview: "35", // Magenta
	types.Completed:   "32", // Green
}

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

const TASK_NAME_LENGTH = 40
func printKanbanHeader() {
	//fmt.Print(" " + strings.Repeat("╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ ", 4) + "\r\n")
	//fmt.Print(KanbanTaskName("Pending") + KanbanTaskName("In Progress") + KanbanTaskName("In Review") + KanbanTaskName("Completed") + "\r\n")
	//fmt.Print(" " + strings.Repeat("├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ ", 4) + "\r\n")
	fmt.Print(genKanbanHeader())
}

func genKanbanHeader() string {
	var header strings.Builder
	// top bars in each color
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮", borderColors[types.Pending]))
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮", borderColors[types.InProgress]))
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮", borderColors[types.NeedsReview]))
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ \n", borderColors[types.Completed]))

	//header.WriteString(" " + strings.Repeat("╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ ", 4) + "\n")
	header.WriteString(KanbanTaskName("To Do", types.Pending) + KanbanTaskName("In Progress", types.InProgress) + KanbanTaskName("In Review", types.NeedsReview) + KanbanTaskName("Completed", types.Completed) + "\n")

	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤", borderColors[types.Pending]))
	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤", borderColors[types.InProgress]))
	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤", borderColors[types.NeedsReview]))
	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ \n", borderColors[types.Completed]))
	//
	//header.WriteString(" " + strings.Repeat("├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ ", 4) + "\n")
	return header.String()
}

func printKanbanFooter() {
	fmt.Print(genKanbanFooter())
}

func genKanbanFooter() string {
	builder := strings.Builder{}
	// bottom bars in each color
	for status := types.Pending; status <= types.Completed; status++ {
		builder.WriteString(utils.ColoredString(" ╰" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╯", borderColors[status]))
	}
	return builder.String()
}

func BorderColorFromString(status string) string {
	switch status {
	case "Pending":
		return borderColors[types.Pending]
	case "In Progress":
		return borderColors[types.InProgress]
	case "In Review":
		return borderColors[types.NeedsReview]
	case "Completed":
		return borderColors[types.Completed]
	default:
		return "34" // Default to blue
	}
}

func KanbanTaskName(name string, status types.Status ) string {
	return utils.LeftRightBorderedString(name, TASK_NAME_LENGTH, len(name), true, borderColors[status])
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
				line.WriteString(KanbanTaskName("", status))
				continue;
			}
			line.WriteString(KanbanTaskName(taskLists[status][i].Name, status))
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
				line.WriteString(KanbanTaskName("", status))
				continue;
			}
			line.WriteString(KanbanTaskName(taskLists[status][i].Name, status))
		}
		builder.WriteString(line.String() + " \n")
	}
	builder.WriteString(genKanbanFooter())
	return builder.String()
}
