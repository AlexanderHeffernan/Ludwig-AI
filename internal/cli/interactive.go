package cli

import (
	"fmt"
	"os"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"

	"ludwig/internal/storage"
	"ludwig/internal/types/model"
)

// StartInteractive runs the interactive bubbletea UI.
func StartInteractive(version string) {
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing task storage: %v\n", err)
		os.Exit(1)
	}

	m := model.NewModel(taskStore, version)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// stripAnsiCodes removes ANSI escape sequences from a string to get the visible length
func stripAnsiCodes(s string) string {
	// Remove ANSI escape sequences (cursor movement, colors, etc.)
	ansiRegex := `\x1b\[[0-9;]*[a-zA-Z]`
	re := regexp.MustCompile(ansiRegex)
	return re.ReplaceAllString(s, "")
}
