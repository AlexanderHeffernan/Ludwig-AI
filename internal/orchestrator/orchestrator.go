package orchestrator

import (
	"log"
	"sync"
	"time"

	"ludwig/internal/orchestrator/clients"
	"ludwig/internal/storage"
	"ludwig/internal/types"
)

var (
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
)

// Start launches the orchestrator loop in a goroutine.
func Start() {
	mu.Lock()
	defer mu.Unlock()
	if running {
		return
	}
	running = true
	stopCh = make(chan struct{})
	wg.Add(1)
	go orchestratorLoop()
}

// Stop signals the orchestrator to stop and waits for it to finish.
func Stop() {
	mu.Lock()
	if !running {
		mu.Unlock()
		return
	}
	close(stopCh)
	mu.Unlock()
	wg.Wait()
	mu.Lock()
	running = false
	mu.Unlock()
}

// IsRunning returns true if the orchestrator is running.
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}

// orchestratorLoop processes tasks, polling for new ones.
func orchestratorLoop() {
	defer wg.Done()
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		log.Printf("Failed to initialize task storage: %v", err)
		return
	}
	gemini := &clients.GeminiClient{}

	for {
		select {
		case <-stopCh:
			return
		default:
			// Process tasks in order: NeedsReview (with responses), then Pending
			tasks, err := taskStore.ListTasks()
			if err != nil {
				log.Printf("Failed to list tasks: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			processed := false

			// Check for NeedsReview tasks with responses
			for _, t := range tasks {
				if t.Status == types.NeedsReview && t.ReviewResponse != nil {
					log.Printf("Resuming task %s with user response: %s", t.ID, t.ReviewResponse.ChosenLabel)
					t.Status = types.InProgress
					if err := taskStore.UpdateTask(t); err != nil {
						log.Printf("Failed to set task %s to In Progress: %v", t.ID, err)
						continue
					}

					optionLabels := make([]string, len(t.Review.Options))
					for i, opt := range t.Review.Options {
						optionLabels[i] = opt.Label
					}
					prompt := BuildResumePrompt(t.Name, t.WorkInProgress, t.Review.Question, optionLabels, t.ReviewResponse.ChosenLabel, t.ReviewResponse.UserNotes)
					response, err := gemini.SendPrompt(prompt)
					if err != nil {
						log.Printf("Error resuming task %s: %v", t.ID, err)
						t.Status = types.NeedsReview
						_ = taskStore.UpdateTask(t)
						continue
					}
					log.Printf("Completed task %s: Gemini response: %s", t.ID, response)
					t.Status = types.Completed
					_ = taskStore.UpdateTask(t)
					processed = true
					break
				}
			}

			if processed {
				continue
			}

			// Process pending tasks
			for _, t := range tasks {
				if t.Status == types.Pending {
					log.Printf("Starting task %s: %s", t.ID, t.Name)
					t.Status = types.InProgress
					if err := taskStore.UpdateTask(t); err != nil {
						log.Printf("Failed to set task %s to In Progress: %v", t.ID, err)
						continue
					}
					response, err := gemini.SendPrompt(BuildTaskPrompt(t.Name))
					if err != nil {
						log.Printf("Error sending task %s to Gemini: %v", t.ID, err)
						t.Status = types.Pending
						_ = taskStore.UpdateTask(t)
						continue
					}

					// Check if response contains a review request
					workInProgress, review, hasReview := parseReviewRequest(response)
					if hasReview {
						log.Printf("Task %s needs review: %s", t.ID, review.Question)
						t.Status = types.NeedsReview
						t.WorkInProgress = workInProgress
						t.Review = review
						_ = taskStore.UpdateTask(t)
						processed = true
						break
					}

					log.Printf("Completed task %s: Gemini response: %s", t.ID, response)
					t.Status = types.Completed
					_ = taskStore.UpdateTask(t)
					processed = true
					break // Only process one task per loop
				}
			}
			if !processed {
				log.Printf("No pending tasks found. Waiting before polling again.")
				time.Sleep(2 * time.Second) // No pending tasks, wait before polling again
			}
		}
	}
}



// parseReviewRequest extracts a review request and work-in-progress from the AI response
// Returns (WorkInProgress, ReviewRequest, hasReview)
func parseReviewRequest(response string) (string, *types.ReviewRequest, bool) {
	// Look for NEEDS_REVIEW markers
	reviewStart := -1
	reviewEnd := -1
	for i := 0; i < len(response)-len("---NEEDS_REVIEW---"); i++ {
		if response[i:i+len("---NEEDS_REVIEW---")] == "---NEEDS_REVIEW---" {
			reviewStart = i + len("---NEEDS_REVIEW---")
			break
		}
	}
	if reviewStart == -1 {
		return "", nil, false
	}

	for i := reviewStart; i < len(response)-len("---END_REVIEW---"); i++ {
		if response[i:i+len("---END_REVIEW---")] == "---END_REVIEW---" {
			reviewEnd = i
			break
		}
	}
	if reviewEnd == -1 {
		return "", nil, false
	}

	// Extract work-in-progress (everything before NEEDS_REVIEW marker)
	workInProgress := trim(response[:reviewStart-len("---NEEDS_REVIEW---")])

	reviewBlock := response[reviewStart:reviewEnd]
	review := parseReviewBlock(reviewBlock)
	return workInProgress, review, review != nil
}

// parseReviewBlock parses the content between NEEDS_REVIEW markers
func parseReviewBlock(block string) *types.ReviewRequest {
	lines := split(block, "\n")
	var question, context string
	var options []types.ReviewOption

	for _, line := range lines {
		line = trim(line)
		if hasPrefix(line, "Question:") {
			question = trimPrefix(line, "Question:")
			question = trim(question)
		} else if hasPrefix(line, "Context:") {
			context = trimPrefix(line, "Context:")
			context = trim(context)
		} else if hasPrefix(line, "- id:") {
			opt := parseOption(line)
			if opt != nil {
				options = append(options, *opt)
			}
		}
	}

	if question == "" {
		return nil
	}

	return &types.ReviewRequest{
		Question:  question,
		Options:   options,
		Context:   context,
		CreatedAt: time.Now(),
	}
}

// parseOption extracts an option from "- id: x | label: y" format
func parseOption(line string) *types.ReviewOption {
	// Remove leading "- id: "
	if !hasPrefix(line, "- id:") {
		return nil
	}
	line = trimPrefix(line, "- id:")
	line = trim(line)

	// Split on "|"
	parts := split(line, "|")
	if len(parts) < 2 {
		return nil
	}

	id := trim(parts[0])
	labelPart := trim(parts[1])

	// Extract label from "label: y"
	if !hasPrefix(labelPart, "label:") {
		return nil
	}
	label := trimPrefix(labelPart, "label:")
	label = trim(label)

	return &types.ReviewOption{
		ID:    id,
		Label: label,
	}
}

// Helper string functions
func split(s, sep string) []string {
	var result []string
	for {
		index := indexOf(s, sep)
		if index == -1 {
			result = append(result, s)
			break
		}
		result = append(result, s[:index])
		s = s[index+len(sep):]
	}
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func trimPrefix(s, prefix string) string {
	if hasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
