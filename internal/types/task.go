package types

import (
	"fmt"
	"time"
)

type Status int

const (
	Pending Status = iota
	InProgress
	NeedsReview
	Completed
)

type Task struct {
	ID     string
	Name   string
	Status Status
	CreatedAt   time.Time

	BranchName     string          // Git branch created for this task
	WorkInProgress string          // Stores intermediate work before requesting review
	Review         *ReviewRequest
	ReviewResponse *ReviewResponse
	ResponseFile   string          // Path to file containing AI response stream
}

type ReviewRequest struct {
	Question  string
	Options   []ReviewOption
	Context   string
	CreatedAt time.Time
}

type ReviewOption struct {
	ID    string
	Label string
}

type ReviewResponse struct {
	ChosenOptionID string
	ChosenLabel    string
	UserNotes      string
	RespondedAt    time.Time
}

func StatusString(task Task) string {
	switch task.Status {
	case Pending:
		return "Pending"
	case InProgress:
		return "In Progress"
	case NeedsReview:
		return "In Review"
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

// ExampleTasks creates sample tasks for testing
func ExampleTasks() []Task {
	return []Task{
		{
			ID:     "task-1",
			Name:   "Create user authentication",
			Status: Pending,
		},
		{
			ID:     "task-2",
			Name:   "Setup database schema",
			Status: Pending,
		},
		{
			ID:     "task-3",
			Name:   "Design API endpoints",
			Status: Pending,
		},
	}
}
