package types

import (
	"fmt"
)

type Status int

const (
	Pending Status = iota
	InProgress
	Completed
)

type Task struct {
	ID     string
	Name   string
	Status Status
}

func ExampleTasks() []Task {
	return []Task{
		{"0", "Task 1", Pending},
		{"1", "Task 2", InProgress},
		{"2", "Task 3", Completed},
	}
}

func StatusString(task Task) string {
	switch task.Status {
	case Pending:
		return "Pending"
	case InProgress:
		return "In Progress"
	case Completed:
		return "Completed"
	default:
		return "Unknown"
	}
}

func PrintTasks(tasks []Task) {
	for _, task := range tasks {
		fmt.Println("Task: " + task.Name + ", Status: " + StatusString(task))
	}
}
