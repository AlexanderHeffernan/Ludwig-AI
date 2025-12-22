package utils

import "ludwig/internal/types"

func PointerSliceToValueSlice(pointers []*types.Task) []types.Task {
    if pointers == nil {
        return nil
    }

    values := make([]types.Task, len(pointers))
    for i, ptr := range pointers {
        if ptr != nil {
            values[i] = *ptr  // dereference the pointer
        }
        // If ptr is nil, values[i] will be the zero value of types.Task
    }
    return values
}
