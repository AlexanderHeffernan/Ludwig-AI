package utils

import (
	"ludwig/internal/types/task"
	"sort"
)

func PointerSliceToValueSlice(pointers []*task.Task) []task.Task {
    if pointers == nil {
        return nil
    }

    values := make([]task.Task, len(pointers))
    for i, ptr := range pointers {
        if ptr != nil {
            values[i] = *ptr  // dereference the pointer
        }
        // If ptr is nil, values[i] will be the zero value of task.Task
    }
    sort.Slice(values, func(i, j int) bool {
		return TaskComparator(&values[i], &values[j])
	})
	return values
}

func TaskComparator(a, b *task.Task) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.CreatedAt != b.CreatedAt {
		return a.CreatedAt.Before(b.CreatedAt)
	}
	if len(a.Name) != len(b.Name) {
		return len(a.Name) < len(b.Name)
	}
	return a.ID < b.ID
}
