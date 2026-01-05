package model
import (
	"ludwig/internal/storage"
	"ludwig/internal/types/task"
	"ludwig/internal/utils"
	"ludwig/internal/kanban"
	"ludwig/internal/components/outputViewport"
	"ludwig/internal/components/commandInput"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"time"
	"fmt"
)

type Model struct {
	taskStore *storage.FileTaskStorage
	tasks     []task.Task
	textInput textarea.Model
	commandInput commandInput.Model
	commands  []Command
	err       error
	message   string
	taskViewport outputViewport.Model
	viewingViewport bool


	/*
	spinner   spinner.Model
	viewport  viewport.Model
	progressBar progressBar.Model
	filePath  string
	viewingTask *task.Task
	fileChangeInfo *utils.FileChangeInfo
	*/
}

type Command struct {
	Text string
	Action func(Text string, m *Model) string
	Description string
}

// tickMsg is a message sent on a timer to trigger a refresh.
type tickMsg time.Time

var loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = loadingStyle

	m := &Model{
		taskStore: taskStore,
		tasks:     utils.PointerSliceToValueSlice(tasks),
		commandInput: commandInput.NewModel(),
	}
	m.commands = PalleteCommands(taskStore)
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.taskViewport.Init(),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// Update handles incoming messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	//m.textInput, cmd = m.textInput.Update(msg)
	m.commandInput.Update(msg)
	_, cmd = m.taskViewport.Update(msg)
	// Dynamically adjust height based on content wrapping

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if !m.viewingViewport { return m, tea.Quit }
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
			m.err = fmt.Errorf("command not found: %q", commandText)
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

	return m, cmd
}

const VIEWPORT_CONTROLS = "\n(Press Ctrl+S to scroll down, Ctrl+W to scroll up, Esc to exit view)"

// getScrollbarChars generates scrollbar characters for each line based on viewport state
// View renders the UI.
func (m *Model) View() string {
	var s strings.Builder
	if m.viewingViewport {
		return m.taskViewport.View()
		/*
		s.WriteString(m.progressBar.View())
		// Render full screen output view
		bubbleStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Width(utils.TermWidth() - 5).
			Height(utils.TermHeight() - 8).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Margin(1, 1)

		spinnerOn := m.viewingTask.Status == task.InProgress && orchestrator.IsRunning()

		insideBubble := strings.Builder{}
		insideBubble.WriteString(m.viewport.View())
		if spinnerOn {
			utils.DebugLog("Rendering spinner in viewport view")
			insideBubble.WriteString("\n" + m.spinner.View() + loadingStyle.Render(" Working on it"))
		}

		s.WriteString(bubbleStyle.Render(insideBubble.String()))
		s.WriteString(VIEWPORT_CONTROLS)
		return s.String()
		*/
	}
	// Render the Kanban board.
	s.WriteString(kanban.RenderKanban(m.tasks))
	//s.WriteString("\n")

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

	return s.String()
}


func (m *Model) UpdateTasks() {
	tasks, err := m.taskStore.ListTasks()
	if err != nil {
		m.err = err
	} else {
		m.tasks = utils.PointerSliceToValueSlice(tasks)
	}

	if m.taskViewport.ViewingTask == nil { return }
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

