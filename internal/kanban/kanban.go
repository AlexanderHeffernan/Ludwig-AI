package kanban

import (
	"fmt"
	"strings"
	"ludwig/internal/types/task"
	"ludwig/internal/utils"
	"strconv"
	"slices"
)

var borderColors map[task.Status]string = map[task.Status]string {
	task.Pending:     "34", // Blue
	task.InProgress:  "33", // Yellow
	task.NeedsReview: "35", // Magenta
	task.Completed:   "32", // Green
}

func seperateTaskByStatus(tasks []task.Task) map[task.Status][]task.Task {
	taskLists := map[task.Status][]task.Task{
		task.Pending:     {},
		task.InProgress:  {},
		task.NeedsReview: {},
		task.Completed:   {},
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
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮", borderColors[task.Pending]))
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮", borderColors[task.InProgress]))
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮", borderColors[task.NeedsReview]))
	header.WriteString(utils.ColoredString(" ╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ \n", borderColors[task.Completed]))

	//header.WriteString(" " + strings.Repeat("╭" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╮ ", 4) + "\n")
	header.WriteString(KanbanTaskName("To Do", task.Pending) + KanbanTaskName("In Progress", task.InProgress) + KanbanTaskName("In Review", task.NeedsReview) + KanbanTaskName("Completed", task.Completed) + "\n")

	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤", borderColors[task.Pending]))
	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤", borderColors[task.InProgress]))
	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤", borderColors[task.NeedsReview]))
	header.WriteString(utils.ColoredString(" ├" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "┤ \n", borderColors[task.Completed]))
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
	for status := task.Pending; status <= task.Completed; status++ {
		builder.WriteString(utils.ColoredString(" ╰" + strings.Repeat("─", TASK_NAME_LENGTH - 3) + "╯", borderColors[status]))
	}
	return builder.String()
}

func BorderColorFromString(status string) string {
	switch status {
	case "Pending":
		return borderColors[task.Pending]
	case "In Progress":
		return borderColors[task.InProgress]
	case "In Review":
		return borderColors[task.NeedsReview]
	case "Completed":
		return borderColors[task.Completed]
	default:
		return "34" // Default to blue
	}
}

func KanbanTaskName(name string, status task.Status ) string {
	return utils.LeftRightBorderedString(name, TASK_NAME_LENGTH, len(name), true, borderColors[status])
}

func DisplayKanban(tasks []task.Task) {
	utils.ClearScreen()
	printKanbanHeader()
	taskLists := seperateTaskByStatus(tasks)

	listLengths := []int{
		len(taskLists[task.Pending]),
		len(taskLists[task.InProgress]),
		len(taskLists[task.NeedsReview]),
		len(taskLists[task.Completed]),
	}
	maxListLength := 0
	for _, length := range listLengths {
		if length > maxListLength {
			maxListLength = length
		}
	}

	index := 0
	for i := 0; i < maxListLength; i++ {
		var line strings.Builder
		for status := task.Pending; status <= task.Completed; status++ {
			if i >= len(taskLists[status]) {
				line.WriteString(KanbanTaskName("", status))
				continue;
			}
			displayText := strconv.Itoa(index) + " " + taskLists[status][i].Name
			line.WriteString(KanbanTaskName(displayText, status))
		}
		fmt.Print(line.String() + " \n")
		index++
	}
	printKanbanFooter()
}

func RenderKanban(tasks []task.Task) string {
	var builder strings.Builder
	//printKanbanHeader()
	builder.WriteString(genKanbanHeader())
	taskLists := seperateTaskByStatus(tasks)

	listLengths := []int{
		len(taskLists[task.Pending]),
		len(taskLists[task.InProgress]),
		len(taskLists[task.NeedsReview]),
		len(taskLists[task.Completed]),
	}
	maxListLength := 0
	for _, length := range listLengths {
		if length > maxListLength {
			maxListLength = length
		}
	}

	index := 1
	for i := 0; i < maxListLength; i++ {
		var line strings.Builder
		for status := task.Pending; status <= task.Completed; status++ {
			if i >= len(taskLists[status]) {
				line.WriteString(KanbanTaskName("", status))
				continue;
			}
			task := taskLists[status][i]
			displayText := "#" + strconv.Itoa(slices.Index(tasks, task)) + " " + task.Name
			index++
			line.WriteString(KanbanTaskName(displayText, status))
		}
		builder.WriteString(line.String() + " \n")

	}
	builder.WriteString(genKanbanFooter())
	return builder.String()
}
