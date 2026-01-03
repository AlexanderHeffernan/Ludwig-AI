package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// CreateWorktree creates a new git worktree for a given branch
// Returns the path to the worktree directory
func CreateWorktree(branchName, taskID string) (string, error) {
	repoRoot := getRepoRoot()
	worktreeDir := filepath.Join(repoRoot, ".worktrees", taskID)
	
	// Ensure .worktrees directory exists
	if err := os.MkdirAll(filepath.Join(repoRoot, ".worktrees"), 0755); err != nil {
		return "", fmt.Errorf("failed to create .worktrees directory: %w", err)
	}
	
	// Try to create worktree based on "main" branch first
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreeDir, "main")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err == nil {
		return worktreeDir, nil
	}
	
	// If "main" doesn't exist, try the current branch as fallback
	currentBranch, err := getCurrentBranch(repoRoot)
	if err != nil {
		return "", fmt.Errorf("failed to create worktree: main branch not found and unable to determine current branch")
	}
	
	cmd = exec.Command("git", "worktree", "add", "-b", branchName, worktreeDir, currentBranch)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}
	
	return worktreeDir, nil
}

// getCurrentBranch returns the current branch name or HEAD ref
func getCurrentBranch(repoRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RemoveWorktree removes a git worktree and cleans up the directory
func RemoveWorktree(worktreePath string) error {
	repoRoot := getRepoRoot()
	
	// Remove the worktree
	cmd := exec.Command("git", "worktree", "remove", "-f", worktreePath)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	
	// Clean up any remaining files
	if err := os.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to clean up worktree directory: %w", err)
	}
	
	return nil
}

// CommitAnyChanges stages and commits any uncommitted changes in the worktree
// This ensures that AI work is preserved even if the AI didn't explicitly commit
// Uses the task ID to create a descriptive commit message
func CommitAnyChanges(worktreePath string, taskID string) error {
	// Check if there are any changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	
	// If no changes, nothing to commit
	if len(output) == 0 {
		return nil
	}
	
	// Stage all changes
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = worktreePath
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	
	// Commit the changes
	commitMsg := fmt.Sprintf("Task completed: %s\n\nAuto-committed any uncommitted changes to preserve work.", taskID)
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = worktreePath
	if err := commitCmd.Run(); err != nil {
		// Commit might fail if there are no staged changes after add, which is fine
		return nil
	}
	
	return nil
}

// CreateBranch creates a new branch and checks it out (deprecated: use CreateWorktree instead)
func CreateBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = getRepoRoot()
	return cmd.Run()
}

// CheckoutBranch switches to an existing branch (deprecated: no longer needed with worktrees)
func CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = getRepoRoot()
	return cmd.Run()
}

// BranchExists checks if a branch already exists
func BranchExists(branchName string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	cmd.Dir = getRepoRoot()
	cmd.Stderr = nil
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	// exit code 1 means branch doesn't exist, which is expected
	return false, nil
}

// GenerateBranchName creates a unique branch name from a task name
// Extracts first 2-3 words, converts to kebab-case, ensures uniqueness
func GenerateBranchName(taskName string) (string, error) {
	// Extract first 2-3 meaningful words
	words := extractWords(taskName)
	if len(words) == 0 {
		return "", fmt.Errorf("unable to generate branch name from task: %s", taskName)
	}

	// Take first 2-3 words
	maxWords := 3
	if len(words) < maxWords {
		maxWords = len(words)
	}
	baseName := strings.Join(words[:maxWords], "-")

	// Convert to kebab-case and limit length
	baseName = toKebabCase(baseName)
	if len(baseName) > 40 {
		baseName = baseName[:40]
	}

	// Add ludwig/ prefix
	baseName = "ludwig/" + baseName

	// Check for duplicates and append counter if needed
	branchName := baseName
	counter := 1
	for {
		exists, err := BranchExists(branchName)
		if err != nil {
			return "", err
		}
		if !exists {
			return branchName, nil
		}
		branchName = fmt.Sprintf("%s-%d", baseName, counter)
		counter++
	}
}

// extractWords extracts meaningful words from a task name
func extractWords(taskName string) []string {
	// Split by spaces and special chars
	re := regexp.MustCompile("[^a-zA-Z0-9]+")
	parts := re.Split(taskName, -1)

	var words []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && len(part) > 2 { // Skip single letters and numbers only
			words = append(words, strings.ToLower(part))
		}
	}
	return words
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	// Replace underscores and spaces with hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric except hyphens
	re := regexp.MustCompile("[^a-z0-9-]")
	s = re.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	re = regexp.MustCompile("-+")
	s = re.ReplaceAllString(s, "-")

	// Trim hyphens from start/end
	s = strings.Trim(s, "-")

	return s
}

// getRepoRoot returns the root directory of the repository
func getRepoRoot() string {
	// For now, assume current working directory is within the repo
	// In production, you might want to use git to find the root
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
