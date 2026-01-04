package model
import (
	"ludwig/internal/storage"
	"ludwig/internal/types/task"
	"ludwig/internal/utils"
	"ludwig/internal/kanban"
	"ludwig/internal/types/progressBar"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"time"
	"ludwig/internal/orchestrator"
	"fmt"
	"strconv"
	"math"
)

type Model struct {
	taskStore *storage.FileTaskStorage
	tasks     []task.Task
	textInput textarea.Model
	commands  []Command
	err       error
	message   string
	spinner   spinner.Model
	viewport  viewport.Model
	viewingViewport bool
	progressBar progressBar.Model
	filePath  string
	viewingTask *task.Task
	fileChangeInfo *utils.FileChangeInfo
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
		textInput: ti,
		spinner:   s,
	}
	m.commands = PalleteCommands(taskStore)
	return m
}

func (m *Model) Init() tea.Cmd {
	/*
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
	return tickMsg(t)
	})
	*/
	return tea.Batch(
		m.spinner.Tick, // This keeps the spinner animating (~10-15 FPS by default)
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// Update handles incoming messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	//m.textInput, cmd = m.textInput.Update(msg)

	m.textInput, cmd = m.textInput.Update(msg)
	m.progressBar.Update(msg)
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
		UpdateViewportWidth(m)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.viewingViewport {
				// Exit full screen output view
				m.viewport = viewport.Model{}
				m.viewingViewport = false
				m.fileChangeInfo = nil  // Clean up file change detection
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
						output := cmd.Action(strings.Join(parts, " "), m)
						if parts[0] == "view" {
							//m.viewingTask = utils.GetTaskByPath(m.tasks, m.filePath)
							utils.DebugLog(m.viewingTask.Name)
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
		m.UpdateTasks()
		// Return a new tick command to continue polling.
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case spinner.TickMsg:  // ← ADD THIS
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.MouseMsg:
		if m.viewingViewport {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}

	return m, cmd
}

const VIEWPORT_CONTROLS = "\n(Press Ctrl+S to scroll down, Ctrl+W to scroll up, Esc to exit view)"

// getScrollbarChars generates scrollbar characters for each line based on viewport state
func (m *Model) getScrollbarChars(numLines int) []string {
	if m.viewport.TotalLineCount() <= m.viewport.Height {
		// No scrollbar needed if content fits entirely
		result := make([]string, numLines)
		for i := range result {
			result[i] = " "
		}
		return result
	}

	// Calculate scrollbar properties
	totalLines := float64(m.viewport.TotalLineCount())
	visibleLines := float64(m.viewport.Height)
	scrollbarHeight := math.Max(1, visibleLines/totalLines*float64(numLines))
	scrollbarTop := (float64(m.viewport.YOffset) / (totalLines - visibleLines)) * (float64(numLines) - scrollbarHeight)

	var scrollbarChars []string
	for i := 0; i < numLines; i++ {
		if float64(i) >= scrollbarTop && float64(i) < scrollbarTop+scrollbarHeight {
			scrollbarChars = append(scrollbarChars, "█")
		} else {
			scrollbarChars = append(scrollbarChars, "│")
		}
	}

	return scrollbarChars
}

// renderViewportWithScrollbar renders the viewport content combined with a scrollbar
func (m *Model) renderViewportWithScrollbar() string {
	viewportContent := m.viewport.View()
	contentLines := strings.Split(viewportContent, "\n")
	
	// Get scrollbar characters for each line
	scrollbarChars := m.getScrollbarChars(len(contentLines))
	
	var result strings.Builder
	for i, contentLine := range contentLines {
		// Pad the content line to the viewport width to ensure scrollbar appears at the right edge
		paddedLine := contentLine
		currentWidth := len([]rune(contentLine)) // Use runes to handle unicode properly
		viewportWidth := m.viewport.Width
		
		// Pad with spaces if the line is shorter than viewport width
		if currentWidth < viewportWidth {
			paddedLine += strings.Repeat(" ", viewportWidth - currentWidth - 1)
		}
		
		// Add scrollbar character at the end
		if i < len(scrollbarChars) {
			paddedLine += scrollbarChars[i]
		}
		
		result.WriteString(paddedLine)
		
		// Add newline except for the last line
		if i < len(contentLines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}
// View renders the UI.
func (m *Model) View() string {
	var s strings.Builder
	if m.viewingViewport {
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
	}
	// Render the Kanban board.
	s.WriteString(kanban.RenderKanban(m.tasks))
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

func (m *Model) ViewportUpdateLoop()  {
	time.AfterFunc(2*time.Second, func() {
		if !m.viewingViewport || m.fileChangeInfo == nil {
			return
		}
		
		changed, fileContent, err := utils.HasFileChangedHybrid(m.filePath, m.fileChangeInfo)
		if err != nil {
			// Handle error, maybe retry or log
			m.ViewportUpdateLoop()
			return
		}
		
		if !changed {
			m.ViewportUpdateLoop()
			return
		}
		
		scrollPrcnt := m.viewport.ScrollPercent()
		utils.DebugLog(strconv.FormatFloat(scrollPrcnt, 'f', -1, 64))
		atBottom := scrollPrcnt > 0.95
		content := utils.OutputLines(strings.Split(fileContent, "\n"))
		m.viewport.SetContent(content)
		if atBottom {
			m.viewport.GotoBottom()
		}
		m.ViewportUpdateLoop()
	})
}

func (m *Model) UpdateTasks() {
	tasks, err := m.taskStore.ListTasks()
	if err != nil {
		m.err = err
	} else {
		m.tasks = utils.PointerSliceToValueSlice(tasks)
	}

	if m.viewingTask == nil { return }
	// Refresh the viewing task details if in viewport mode
	updatedTask, err := m.taskStore.GetTask(m.viewingTask.ID)
	if err != nil {
		m.err = err
		return
	}
	if updatedTask != nil {
		m.viewingTask = updatedTask
	}

}

func UpdateViewportWidth(m *Model) {
	termWidth := utils.TermWidth()
	termHeight := utils.TermHeight()
	if m.viewingViewport {
		m.viewport.Width = termWidth - 14
		m.viewport.Height = termHeight - 6
	}
}
