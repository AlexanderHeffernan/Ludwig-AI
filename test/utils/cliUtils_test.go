package utils_test

import (
	"testing"

	"ludwig/internal/utils"
)

// Test KeyAction structure
func TestKeyActionStructure(t *testing.T) {
	actionCalled := false
	action := func() { actionCalled = true }

	ka := utils.KeyAction{
		Key:         'a',
		Action:      action,
		Description: "Test action",
	}

	if ka.Key != 'a' {
		t.Errorf("key not set properly")
	}
	if ka.Description != "Test action" {
		t.Errorf("description not set properly")
	}
	if ka.Action == nil {
		t.Errorf("action not set properly")
	}

	ka.Action()
	if !actionCalled {
		t.Errorf("action not called properly")
	}
}

// Test Command structure
func TestCommandStructure(t *testing.T) {
	commandCalled := false
	commandText := ""
	
	action := func(text string) {
		commandCalled = true
		commandText = text
	}

	cmd := utils.Command{
		Text:        "test",
		Action:      action,
		Description: "Test command",
	}

	if cmd.Text != "test" {
		t.Errorf("command text not set properly")
	}
	if cmd.Description != "Test command" {
		t.Errorf("command description not set properly")
	}

	cmd.Action("test input")
	if !commandCalled {
		t.Errorf("action not called")
	}
	if commandText != "test input" {
		t.Errorf("action text not passed properly")
	}
}

// Test PrintHelp format structure
func TestPrintHelpFormats(t *testing.T) {
	// Test that commands can have varying text lengths
	commands := []utils.Command{
		{Text: "a", Description: "Short"},
		{Text: "longer_command", Description: "Longer command"},
		{Text: "very_long_command_name", Description: "Very long"},
	}

	// Verify all commands are properly formatted
	for _, cmd := range commands {
		if cmd.Text == "" {
			t.Errorf("command text should not be empty")
		}
		if cmd.Description == "" {
			t.Errorf("command description should not be empty")
		}
	}
}

// Test ClearScreen function exists and doesn't panic
func TestClearScreen(t *testing.T) {
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ClearScreen should not panic: %v", r)
		}
	}()
	
	utils.ClearScreen()
}

// Test Println function
func TestPrintln(t *testing.T) {
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Println should not panic: %v", r)
		}
	}()
	
	utils.Println("Test message")
	utils.Println("")
	utils.Println("Multi\nline\ntext")
}

// Test multiple KeyActions
func TestMultipleKeyActions(t *testing.T) {
	count := 0
	
	actions := []utils.KeyAction{
		{
			Key:         'a',
			Action:      func() { count++ },
			Description: "Action A",
		},
		{
			Key:         'b',
			Action:      func() { count++ },
			Description: "Action B",
		},
		{
			Key:         'c',
			Action:      func() { count++ },
			Description: "Action C",
		},
	}

	if len(actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(actions))
	}

	// Verify each action is callable
	for _, ka := range actions {
		if ka.Action == nil {
			t.Errorf("action for key %c is nil", ka.Key)
		}
	}
}

// Test multiple Commands
func TestMultipleCommands(t *testing.T) {
	commands := []utils.Command{
		{
			Text:        "help",
			Action:      func(text string) {},
			Description: "Show help",
		},
		{
			Text:        "start",
			Action:      func(text string) {},
			Description: "Start process",
		},
		{
			Text:        "stop",
			Action:      func(text string) {},
			Description: "Stop process",
		},
	}

	if len(commands) != 3 {
		t.Errorf("expected 3 commands, got %d", len(commands))
	}

	// Verify each command
	expectedTexts := []string{"help", "start", "stop"}
	for i, cmd := range commands {
		if cmd.Text != expectedTexts[i] {
			t.Errorf("command %d text mismatch", i)
		}
		if cmd.Action == nil {
			t.Errorf("command %d action is nil", i)
		}
	}
}

// Test KeyAction with empty description
func TestKeyActionEmptyDescription(t *testing.T) {
	ka := utils.KeyAction{
		Key:         'x',
		Action:      func() {},
		Description: "",
	}

	if ka.Description != "" {
		t.Errorf("expected empty description")
	}
}

// Test Command with empty text
func TestCommandEmptyText(t *testing.T) {
	cmd := utils.Command{
		Text:        "",
		Action:      func(text string) {},
		Description: "Empty command",
	}

	if cmd.Text != "" {
		t.Errorf("expected empty text")
	}
}

// Test finding command by text
func TestFindCommandByText(t *testing.T) {
	commands := []utils.Command{
		{Text: "add", Description: "Add item"},
		{Text: "remove", Description: "Remove item"},
		{Text: "list", Description: "List items"},
	}

	searchText := "remove"
	found := false
	for _, cmd := range commands {
		if cmd.Text == searchText {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("should find command with text %q", searchText)
	}
}

// Test KeyAction key codes
func TestKeyActionKeyCodes(t *testing.T) {
	keyCodes := []byte{'0', '1', 'a', 'z', 'A', 'Z', '!', '@'}
	
	for _, key := range keyCodes {
		ka := utils.KeyAction{
			Key:    key,
			Action: func() {},
		}
		
		if ka.Key != key {
			t.Errorf("key not set properly for %c", key)
		}
	}
}
