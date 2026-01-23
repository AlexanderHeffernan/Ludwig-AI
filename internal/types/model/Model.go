package model

import (
	"ludwig/internal/components/commandInput"
	"ludwig/internal/components/outputViewport"
	"ludwig/internal/components/orchestratorIndicator"
	"ludwig/internal/kanban"
	"ludwig/internal/storage"
	"ludwig/internal/types/task"
	"ludwig/internal/updater"
	"ludwig/internal/utils"

	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	taskStore       *storage.FileTaskStorage
	tasks           []task.Task
	textInput       textarea.Model
	commandInput    commandInput.Model
	commands        []Command
	err             error
	message         string
	taskViewport    outputViewport.Model
	viewingViewport bool
	orchestratorIndicator *orchestratorIndicator.Model
}

type Command struct {
	Text        string
	Action      func(Text string, m *Model) string
	Description string
}

// tickMsg is a message sent on a timer to trigger a refresh.
type tickMsg time.Time

var loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

func NewModel(taskStore *storage.FileTaskStorage, version string) *Model {
	ti := textarea.New()
	ti.Placeholder = "...Enter command (e.g., 'add <task>', 'exit', 'help')"

	// account for border padding
	ti.SetWidth(utils.TermWidth() - 6)

	// Start with minimum height
	ti.SetHeight(2)
	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ti.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ti.ShowLineNumbers = false
	ti.Prompt = ""

	// setting to 0 removes the character limit
	ti.CharLimit = 0
	ti.Focus()

	tasks, err := taskStore.ListTasks()
	if err != nil {
		// This error will be displayed in the view.
		return &Model{err: fmt.Errorf("could not load tasks: %w", err)}
	}

	m := &Model{
		taskStore:    taskStore,
		tasks:        utils.PointerSliceToValueSlice(tasks),
		commandInput: commandInput.NewModel(),
		taskViewport: outputViewport.NewModel(),
		orchestratorIndicator: orchestratorIndicator.NewModel(),
	}
	m.commands = PalleteCommands(taskStore)

	m.checkForUpdate(version)

	return m
}

func (m *Model) checkForUpdate(version string) {
	// Check for updates in the background
	go func() {
		isNewer, latestVersion, err := updater.CheckForUpdate(version)
		if err == nil && isNewer {
			m.message = fmt.Sprintf("Update available: %s → %s. Exit Ludwig and run 'ludwig --update' to install.", version, latestVersion)
		}
	}()
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.taskViewport.Init(),
		m.orchestratorIndicator.Init(),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// Update handles incoming messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	var cmds []tea.Cmd

	if !m.viewingViewport {
		var inputCmd tea.Cmd
		m.commandInput, inputCmd = m.commandInput.Update(msg)
		if inputCmd != nil {
			cmds = append(cmds, inputCmd)
		}
	}
	var indicatorCmd tea.Cmd
	m.orchestratorIndicator, indicatorCmd = m.orchestratorIndicator.Update(msg)
	if indicatorCmd != nil {
		cmds = append(cmds, indicatorCmd)
	}
	_, viewportCmd := m.taskViewport.Update(msg)
	if viewportCmd != nil {
		cmds = append(cmds, viewportCmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if !m.viewingViewport {
				return m, tea.Quit
			}
			m.viewingViewport = false
			return m, nil
		case tea.KeyEnter:
			input := strings.TrimSpace(m.commandInput.TextInput.Value())
			parts := strings.Fields(input)
			m.commandInput.TextInput.SetValue("")
			m.err = nil

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
						output := cmd.Action(strings.Join(parts, " "), m)
						if parts[0] != "view" {
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
			//m.err = fmt.Errorf("command not found: %q", commandText)
			m.message = "Command not found: " + parts[0]
			return m, nil
		}

	case tickMsg:
		// On each tick, reload tasks from storage.
		m.UpdateTasks()
		// Return a new tick command to continue polling.
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case error:
		m.err = msg
		return m, nil
	}

	// Handle commands from viewport (like spinner ticks)
	/*
	if viewportCmd != nil {
		if cmd != nil {
			return m, tea.Batch(cmd, viewportCmd)
		}
		return m, viewportCmd
	}

	return m, cmd
	*/
	return m, tea.Batch(cmds...)
}

const VIEWPORT_CONTROLS = "\n(Press Ctrl+S to scroll down, Ctrl+W to scroll up, Esc to exit view)"

// getScrollbarChars generates scrollbar characters for each line based on viewport state
// View renders the UI.
func (m *Model) View() string {
	var s strings.Builder
	if m.viewingViewport {
		return m.taskViewport.View()
	}
	// Render the Kanban board.
	s.WriteString(kanban.RenderKanban(m.tasks))

	linesCount := strings.Count(s.String(), "\n")

	padStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Height(utils.TermHeight() - linesCount - m.commandInput.Height - 3).
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

	s.WriteString(m.commandInput.View())

	//if orchestrator.IsRunning() {
		//s.WriteString(loadingStyle.Render("\nOrchestrator is running..."))
	//}
	s.WriteString(m.orchestratorIndicator.View())

	return s.String()
}

func (m *Model) UpdateTasks() {
	tasks, err := m.taskStore.ListTasks()
	if err != nil {
		m.err = err
	} else {
		m.tasks = utils.PointerSliceToValueSlice(tasks)
	}

	if m.taskViewport.ViewingTask == nil {
		return
	}
	// Refresh the viewing task details if in viewport mode
	updatedTask, err := m.taskStore.GetTask(m.taskViewport.ViewingTask.ID)
	if err != nil {
		m.err = err
		return
	}
	if updatedTask != nil {
		m.taskViewport.ViewingTask = updatedTask
	}

}
