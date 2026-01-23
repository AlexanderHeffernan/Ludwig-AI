package commandInput
import (
	"ludwig/internal/utils"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var BORDER_STYLE = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(0, 1).
	Margin(1, 1).
	MarginBottom(0)

type Model struct {
	TextInput textarea.Model
	Height int
}

func NewModel() Model {
	ti := textarea.New()
	ti.Placeholder = "...Enter command (e.g., 'add <task>', 'exit', 'help')"
	ti.SetWidth(utils.TermWidth() - 6) // Account for border padding
	ti.SetHeight(2)                     // Start with minimum height
	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ti.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.Focus()
	return Model{
		TextInput: ti,
	}
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	content := m.TextInput.Value()
	if content == "" {
		m.TextInput.SetHeight(2)
	} else {
		// Calculate wrapped lines based on textarea width
		width := m.TextInput.Width()
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
				if currentLineLength > width {
					wrappedLines++
					currentLineLength = 1
				}
			}
		}

		if wrappedLines < 1 {
			wrappedLines = 1
		}
		m.TextInput.SetHeight(wrappedLines + 1)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		termWidth := msg.Width
		inputWidth := max(termWidth-6, 20) // Account for border + padding
		m.TextInput.SetWidth(inputWidth)
		return *m, nil
	}

	m.Height = m.TextInput.Height()
	return *m, cmd
}

func (m *Model) View() string {
	termWidth := utils.TermWidth()
	inputWidth := max(termWidth - 6, 20) // Account for border (4) + padding (2)
	m.TextInput.SetWidth(inputWidth)

	// Render the middle of the bubble with the input
	inputText := m.TextInput.View()

	borderStyle := BORDER_STYLE.Width(inputWidth)

	return borderStyle.Render(inputText) + "\n"
}
