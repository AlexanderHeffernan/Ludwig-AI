package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"ludwig/internal/storage"
	"ludwig/internal/types"
	"ludwig/internal/utils"

	//"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"bytes"
)

// StartInteractive runs the interactive bubbletea UI.
func StartInteractive() {
	taskStore, err := storage.NewFileTaskStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing task storage: %v\n", err)
		os.Exit(1)
	}

	m := NewModel(taskStore)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// Model represents the state of the application.
type Model struct {
	taskStore *storage.FileTaskStorage
	tasks     []types.Task
	textInput textarea.Model
	commands  []utils.Command
	err       error
	message   string
	viewport viewport.Model
	filePath string
	viewingViewport bool
	viewingTask *types.Task
}

// tickMsg is a message sent on a timer to trigger a refresh.
type tickMsg time.Time

// stripAnsiCodes removes ANSI escape sequences from a string to get the visible length
func stripAnsiCodes(s string) string {
	// Remove ANSI escape sequences (cursor movement, colors, etc.)
	ansiRegex := `\x1b\[[0-9;]*[a-zA-Z]`
	re := regexp.MustCompile(ansiRegex)
	return re.ReplaceAllString(s, "")
}

// NewModel creates a new model with initial values.
func NewModel(taskStore *storage.FileTaskStorage) *Model {
	ti := textarea.New()
	ti.Placeholder = "...Enter command (e.g., 'add <task>', 'exit', 'help')"
	ti.SetWidth(utils.TermWidth() - 6) // Account for border padding
	ti.SetHeight(2) // Start with minimum height
	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ti.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.CharLimit = 0 // No character limit
	ti.Focus()

	tasks, err := taskStore.ListTasks()
	if err != nil {
		// This error will be displayed in the view.
		return &Model{err: fmt.Errorf("could not load tasks: %w", err)}
	}

	m := &Model{
		taskStore: taskStore,
		tasks:     utils.PointerSliceToValueSlice(tasks),
		textInput: ti,
		message:   "",
		err:       nil,
	}
	m.commands = PalleteCommands(taskStore)
	return m
}


// Init initializes the application with a command to start the timer.
func (m *Model) Init() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles incoming messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	//m.textInput, cmd = m.textInput.Update(msg)

	m.textInput, cmd = m.textInput.Update(msg)
	// Dynamically adjust height based on content wrapping
	content := m.textInput.Value()
	if content == "" {
		m.textInput.SetHeight(2)
	} else {
		// Calculate wrapped lines based on textarea width
		width := m.textInput.Width()
		if width <= 0 {
			width = utils.TermWidth() - 6
		}
		wrappedLines := 1
		currentLineLength := 0
		
		for _, char := range content {
			if char == '\n' {
				wrappedLines++
				currentLineLength = 0
			} else {
				currentLineLength++
				if currentLineLength >= width {
					wrappedLines++
					currentLineLength = 0
				}
			}
		}
		
		if wrappedLines < 1 {
			wrappedLines = 1
		}
		m.textInput.SetHeight(wrappedLines + 1)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle terminal resize
		termWidth := msg.Width
		inputWidth := max(termWidth - 6, 20) // Account for border + padding

		m.textInput.SetWidth(inputWidth)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.viewingViewport {
				// Exit full screen output view
				m.viewport = viewport.Model{}
				m.viewingViewport = false
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textInput.Value())
			parts := strings.Fields(input)
			m.textInput.SetValue("")
			//m.message = "" // Clear previous message
			m.err = nil     // Clear previous error

			if len(parts) == 0 {
				return m, nil
			}

			commandText := parts[0]
			if commandText == "exit" {
				return m, tea.Quit
			}

			for _, cmd := range m.commands {
				if cmd.Text == commandText {
					// Execute the command's action.
					if cmd.Action != nil {
						output := cmd.Action(strings.Join(parts, " "))
						if parts[0] == "view" {
							m.viewport = viewport.New(utils.TermWidth() - 4, utils.TermHeight() - 6)
							m.viewport.SetContent(output)
							m.viewport.GotoBottom()
							m.filePath = strings.SplitN(output, "\n", 2)[0]
							m.viewingTask = utils.GetTaskByPath(m.tasks, m.filePath)
							utils.DebugLog(m.viewingTask.Name)
							m.viewingViewport = true
							m.ViewportUpdateLoop(utils.GetFileHash(m.filePath))
						} else {
							m.message = output
						}
					}
					// After action, refresh tasks immediately.
					tasks, err := m.taskStore.ListTasks()
					if err != nil {
						m.err = err
					} else {
						m.tasks = utils.PointerSliceToValueSlice(tasks)
					}
					return m, nil
				}
			}
			m.err = fmt.Errorf("command not found: %q", commandText)
			return m, nil
		case tea.KeyCtrlS:
			m.viewport.ScrollDown((utils.TermHeight() - 6)/2)
			return m, nil
		case tea.KeyCtrlW:
			m.viewport.ScrollUp((utils.TermHeight() - 6)/2)
			return m, nil
		}

	case tickMsg:
		// On each tick, reload tasks from storage.
		tasks, err := m.taskStore.ListTasks()
		if err != nil {
			m.err = err
		} else {
			if len(tasks) != len(m.tasks) {
				m.tasks = utils.PointerSliceToValueSlice(tasks)
			}
		}
		// Return a new tick command to continue polling.
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case error:
		m.err = msg
		return m, nil
	}

	return m, cmd
}

const VIEWPORT_CONTROLS = "\n(Press Ctrl+S to scroll down, Ctrl+W to scroll up, Esc to exit view)"
// View renders the UI.
func (m *Model) View() string {
	var s strings.Builder
	if m.viewingViewport {
		// Render full screen output view
		bubbleStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Width(utils.TermWidth() - 4).
			Height(utils.TermHeight() - 8).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 1).
			Margin(1, 1)

		s.WriteString(bubbleStyle.Render(m.viewport.View()))
		s.WriteString(VIEWPORT_CONTROLS)
		s.WriteString(utils.ThinkingString(*m.viewingTask))
		return s.String()
	}
	// Render the Kanban board.
	s.WriteString(RenderKanban(m.tasks))
	//s.WriteString("\n")

	linesCount := strings.Count(s.String(), "\n")
	
	padStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Height(utils.TermHeight() - linesCount - m.textInput.Height() - 3).
		MarginBottom(0)
	// Render output messages
	if m.message != "" || m.err != nil {
		// Only add padding when there's actually content to show
		if m.message != "" {
			s.WriteString(padStyle.Render(m.message))
		}

		if m.err != nil {
			s.WriteString(padStyle.Render("Error: " + m.err.Error()))
		}
	} else {
		// Add empty padding to separate Kanban from input
		s.WriteString(padStyle.Render(""))
	}
	
	// Render the text input for commands with bubble border.
	termWidth := utils.TermWidth()
	
	// Update textarea width to match the available space in the border
	inputWidth := max(termWidth - 6, 20) // Account for border (4) + padding (2)
	m.textInput.SetWidth(inputWidth)
	
	// Render the middle of the bubble with the input
	inputText := m.textInput.View()
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(termWidth - 4).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Margin(1, 1)

	s.WriteString(borderStyle.Render(inputText))
	
	return s.String()
}

func (m *Model) ViewportUpdateLoop(lastHash []byte)  {
	time.AfterFunc(2*time.Second, func() {
		if !m.viewingViewport {
			return
		}
		currentHash, fileContent := utils.GetFileContentHash(m.filePath)
		if bytes.Equal(currentHash, lastHash) {
			m.ViewportUpdateLoop(lastHash)
			return
		}
		m.viewport.SetContent(utils.OutputLines(strings.Split(fileContent, "\n")))
		if m.viewport.ScrollPercent() > 0.95 {
			m.viewport.GotoBottom()
		}
		m.ViewportUpdateLoop(currentHash)
	})
}
