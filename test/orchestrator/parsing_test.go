package orchestrator_test

import (
	"testing"

	"ludwig/internal/orchestrator"
)

// Note: Many parsing functions in orchestrator package are private (lowercase).
// The public APIs we can test are:
// - Start/Stop/IsRunning (orchestrator control)
// - GenerateBranchName (git operations)
// - Build prompt functions

func TestOrchestratorIsRunning(t *testing.T) {
	// Test initial state
	if orchestrator.IsRunning() {
		t.Errorf("expected orchestrator to not be running initially")
	}
}

func TestOrchestratorStartStop(t *testing.T) {
	// Clean slate
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	// Start orchestrator
	orchestrator.Start()
	if !orchestrator.IsRunning() {
		t.Errorf("expected orchestrator to be running after Start()")
	}

	// Stop orchestrator
	orchestrator.Stop()
	if orchestrator.IsRunning() {
		t.Errorf("expected orchestrator to not be running after Stop()")
	}
}

func TestOrchestratorMultipleStarts(t *testing.T) {
	// Clean up first
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	// Starting twice should not cause issues
	orchestrator.Start()
	if !orchestrator.IsRunning() {
		t.Errorf("first start failed")
	}

	orchestrator.Start()
	if !orchestrator.IsRunning() {
		t.Errorf("orchestrator should still be running")
	}

	orchestrator.Stop()
}

func TestOrchestratorMultipleStops(t *testing.T) {
	// Clean up
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	// Multiple stops should not cause panic
	orchestrator.Start()
	orchestrator.Stop()
	orchestrator.Stop() // Second stop should be safe
}

func TestOrchestratorStopWhenNotRunning(t *testing.T) {
	// Ensure not running
	if orchestrator.IsRunning() {
		orchestrator.Stop()
	}

	// Stopping when not running should not panic
	orchestrator.Stop()
	if orchestrator.IsRunning() {
		t.Errorf("should not be running")
	}
}
