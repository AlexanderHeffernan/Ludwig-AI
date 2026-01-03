package orchestrator

import (
	"sync"
	"time"

	"ludwig/internal/config"
	"ludwig/internal/orchestrator/clients"
	"ludwig/internal/storage"
	"ludwig/internal/types"
)

var (
	mu                sync.Mutex
	running           bool
	stopCh            chan struct{}
	wg                sync.WaitGroup
	rateLimitMu       sync.Mutex
	lastRequestTime   time.Time
	semaphore         chan struct{} // Limits concurrent tasks to 3
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
	semaphore = make(chan struct{}, 3) // Max 3 parallel tasks
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

// orchestratorLoop polls for tasks and dispatches them to a worker pool.
func orchestratorLoop() {
	defer wg.Done()
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		// Silent failure - orchestrator runs in background
		return
	}
	
	// Load configuration (optional)
	cfg, err := config.LoadConfig()
	if err != nil {
		// Config load failure is non-critical, continue without it
	}

	// Initialize AI client based on configuration
	var aiClient clients.AIClient
	if cfg != nil {
		switch cfg.AIProvider {
		case "ollama":
			aiClient = clients.NewOllamaClient(cfg.OllamaBaseURL, cfg.OllamaModel)
		case "copilot":
			aiClient = clients.NewCopilotClient(cfg.CopilotModel)
		default:
			// Default to Gemini
			aiClient = &clients.GeminiClient{}
		}
	} else {
		// Default to Gemini if no config
		aiClient = &clients.GeminiClient{}
	}

	for {
		select {
		case <-stopCh:
			return
		default:
			// Get all tasks and dispatch available ones
			tasks, err := taskStore.ListTasks()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			foundWork := false

			// First pass: process NeedsReview tasks with responses
			for _, t := range tasks {
				if t.Status == types.NeedsReview && t.ReviewResponse != nil {
					// Try to acquire semaphore slot
					select {
					case semaphore <- struct{}{}:
						foundWork = true
						wg.Add(1)
						go processResumeTask(taskStore, aiClient, cfg, t)
					default:
						// No available slots, continue to next task
					}
				}
			}

			// Second pass: process Pending tasks
			for _, t := range tasks {
				if t.Status == types.Pending {
					// Try to acquire semaphore slot
					select {
					case semaphore <- struct{}{}:
						foundWork = true
						wg.Add(1)
						go processNewTask(taskStore, aiClient, cfg, t)
					default:
						// No available slots, continue to next task
					}
				}
			}

			if !foundWork {
				time.Sleep(2 * time.Second) // No tasks available, wait before polling again
			}
		}
	}
}

// processResumeTask handles a NeedsReview task with a user response.
func processResumeTask(taskStore *storage.FileTaskStorage, aiClient clients.AIClient, cfg *config.Config, t *types.Task) {
	defer wg.Done()
	defer func() { <-semaphore }() // Release semaphore slot

	t.Status = types.InProgress
	if err := taskStore.UpdateTask(t); err != nil {
		return
	}

	optionLabels := make([]string, len(t.Review.Options))
	for i, opt := range t.Review.Options {
		optionLabels[i] = opt.Label
	}
	prompt := BuildResumePrompt(t.Name, t.WorkInProgress, t.Review.Question, optionLabels, t.ReviewResponse.ChosenLabel, t.ReviewResponse.UserNotes)

	// Apply rate limiting before request
	applyRateLimit(cfg)

	// Create response writer for streaming
	respWriter, respPath, err := storage.NewResponseWriter(t.ID)
	if err != nil {
		t.Status = types.NeedsReview
		_ = taskStore.UpdateTask(t)
		return
	}
	defer respWriter.Close()

	// Store response file path immediately so it's available during streaming
	t.ResponseFile = respPath
	if err := taskStore.UpdateTask(t); err != nil {
		// Failure to save path is non-critical
	}

	_, err = aiClient.SendPromptWithDir(prompt, respWriter, t.WorktreePath)
	if err != nil {
		t.Status = types.NeedsReview
		_ = taskStore.UpdateTask(t)
		return
	}

	t.Status = types.Completed
	// ResponseFile already set above when streaming started
	_ = taskStore.UpdateTask(t)

	// Commit any uncommitted work before removing worktree
	if t.WorktreePath != "" {
		_ = CommitAnyChanges(t.WorktreePath, t.ID)
		_ = RemoveWorktree(t.WorktreePath)
		t.WorktreePath = ""
		_ = taskStore.UpdateTask(t)
	}
}

// processNewTask handles a Pending task that needs initial processing.
func processNewTask(taskStore *storage.FileTaskStorage, aiClient clients.AIClient, cfg *config.Config, t *types.Task) {
	defer wg.Done()
	defer func() { <-semaphore }() // Release semaphore slot

	// Generate and create worktree for this task
	branchName, err := GenerateBranchName(t.Name)
	if err != nil {
		return
	}

	worktreePath, err := CreateWorktree(branchName, t.ID)
	if err != nil {
		return
	}
	t.BranchName = branchName
	t.WorktreePath = worktreePath

	t.Status = types.InProgress
	if err := taskStore.UpdateTask(t); err != nil {
		return
	}

	// Apply rate limiting before request
	applyRateLimit(cfg)

	// Create response writer for streaming
	respWriter, respPath, err := storage.NewResponseWriter(t.ID)
	if err != nil {
		t.Status = types.Pending
		_ = taskStore.UpdateTask(t)
		return
	}
	defer respWriter.Close()

	// Store response file path immediately so it's available during streaming
	t.ResponseFile = respPath
	if err := taskStore.UpdateTask(t); err != nil {
		// Failure to save path is non-critical
	}

	response, err := aiClient.SendPromptWithDir(BuildTaskPrompt(t.Name), respWriter, t.WorktreePath)
	if err != nil {
		t.Status = types.Pending
		_ = taskStore.UpdateTask(t)
		return
	}

	// Check if response contains a review request
	workInProgress, review, hasReview := parseReviewRequest(response)
	if hasReview {
		t.Status = types.NeedsReview
		t.WorkInProgress = workInProgress
		t.Review = review
		// ResponseFile already set above when streaming started
		_ = taskStore.UpdateTask(t)
		return
	}

	t.Status = types.Completed
	// ResponseFile already set above when streaming started
	_ = taskStore.UpdateTask(t)

	// Commit any uncommitted work before removing worktree
	if t.WorktreePath != "" {
		_ = CommitAnyChanges(t.WorktreePath, t.ID)
		_ = RemoveWorktree(t.WorktreePath)
		t.WorktreePath = ""
		_ = taskStore.UpdateTask(t)
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

// applyRateLimit waits if necessary based on config rate limits (thread-safe).
func applyRateLimit(cfg *config.Config) {
	if cfg == nil || cfg.DelayMs <= 0 {
		return // No rate limiting configured
	}

	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	now := time.Now()
	timeSinceLastRequest := now.Sub(lastRequestTime)
	delay := time.Duration(cfg.DelayMs) * time.Millisecond

	if timeSinceLastRequest < delay {
		waitTime := delay - timeSinceLastRequest
		time.Sleep(waitTime)
	}

	lastRequestTime = time.Now()
}
