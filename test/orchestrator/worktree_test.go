package orchestrator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ludwig/internal/orchestrator"
)

func TestCreateWorktree(t *testing.T) {
	// Get repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Find git repo root by looking for .git directory
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Skip("not in a git repository")
		}
		repoRoot = parent
	}

	// Create a worktree
	branchName, err := orchestrator.GenerateBranchName("test worktree creation")
	if err != nil {
		t.Fatalf("failed to generate branch name: %v", err)
	}

	taskID := "test-worktree-task"
	worktreePath, err := orchestrator.CreateWorktree(branchName, taskID)
	if err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}

	// Defer cleanup
	defer func() {
		if err := orchestrator.RemoveWorktree(worktreePath); err != nil {
			t.Logf("failed to clean up worktree: %v", err)
		}
	}()

	// Verify worktree directory exists
	if _, err := os.Stat(worktreePath); err != nil {
		t.Errorf("worktree directory does not exist: %v", err)
	}

	// Verify worktree path contains taskID
	if !strings.Contains(worktreePath, taskID) {
		t.Errorf("worktree path should contain task ID %q, got %q", taskID, worktreePath)
	}

	// Verify worktree is a directory
	info, err := os.Stat(worktreePath)
	if err != nil {
		t.Errorf("failed to stat worktree path: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("worktree path should be a directory")
	}
}

func TestRemoveWorktree(t *testing.T) {
	// Get repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Find git repo root
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Skip("not in a git repository")
		}
		repoRoot = parent
	}

	// Create a worktree
	branchName, err := orchestrator.GenerateBranchName("test worktree removal")
	if err != nil {
		t.Fatalf("failed to generate branch name: %v", err)
	}

	taskID := "test-worktree-removal-task"
	worktreePath, err := orchestrator.CreateWorktree(branchName, taskID)
	if err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("worktree was not created: %v", err)
	}

	// Remove it
	if err := orchestrator.RemoveWorktree(worktreePath); err != nil {
		t.Fatalf("failed to remove worktree: %v", err)
	}

	// Verify it no longer exists
	if _, err := os.Stat(worktreePath); err == nil {
		t.Errorf("worktree still exists after removal")
	}
}

func TestWorktreePathStructure(t *testing.T) {
	// Get repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Find git repo root
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Skip("not in a git repository")
		}
		repoRoot = parent
	}

	// Create a worktree
	branchName, err := orchestrator.GenerateBranchName("test path structure")
	if err != nil {
		t.Fatalf("failed to generate branch name: %v", err)
	}

	taskID := "test-path-structure-task"
	worktreePath, err := orchestrator.CreateWorktree(branchName, taskID)
	if err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}

	// Defer cleanup
	defer func() {
		if err := orchestrator.RemoveWorktree(worktreePath); err != nil {
			t.Logf("failed to clean up worktree: %v", err)
		}
	}()

	// Verify path contains .worktrees directory
	if !strings.Contains(worktreePath, ".worktrees") {
		t.Errorf("worktree path should contain .worktrees, got %q", worktreePath)
	}

	// Verify the worktree contains a git directory
	gitDir := filepath.Join(worktreePath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Errorf("worktree should contain .git directory: %v", err)
	}
}

func TestMultipleWorktrees(t *testing.T) {
	// Get repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Find git repo root
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Skip("not in a git repository")
		}
		repoRoot = parent
	}

	// Create multiple worktrees
	var worktreePaths []string
	numWorktrees := 3

	for i := 0; i < numWorktrees; i++ {
		branchName, err := orchestrator.GenerateBranchName("parallel task")
		if err != nil {
			t.Fatalf("failed to generate branch name: %v", err)
		}

		taskID := "parallel-task-" + string(rune(i+'0'))
		worktreePath, err := orchestrator.CreateWorktree(branchName, taskID)
		if err != nil {
			t.Fatalf("failed to create worktree %d: %v", i, err)
		}
		worktreePaths = append(worktreePaths, worktreePath)
	}

	// Defer cleanup
	defer func() {
		for _, path := range worktreePaths {
			if err := orchestrator.RemoveWorktree(path); err != nil {
				t.Logf("failed to clean up worktree %s: %v", path, err)
			}
		}
	}()

	// Verify all worktrees exist
	for i, path := range worktreePaths {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("worktree %d does not exist: %v", i, err)
		}
	}

	// Verify all paths are unique
	for i := 0; i < len(worktreePaths); i++ {
		for j := i + 1; j < len(worktreePaths); j++ {
			if worktreePaths[i] == worktreePaths[j] {
				t.Errorf("worktree paths should be unique: %q and %q", worktreePaths[i], worktreePaths[j])
			}
		}
	}
}
