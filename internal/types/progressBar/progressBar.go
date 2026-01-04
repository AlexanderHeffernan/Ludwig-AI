package progressBar

import (
	"strings"
	"math"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"ludwig/internal/utils"
	"strconv"
)

type Model struct {
	Progress float64
	Width int
	Viewport *viewport.Model
}

func NewModel(v *viewport.Model) Model {
	return Model{
		Progress: v.ScrollPercent(),
		Viewport: v,
	}
}

var barStyle = lipgloss.NewStyle().Bold(true)
var style = lipgloss.NewStyle().Faint(true)

func (m *Model) View() string {
	utils.DebugLog("ProgressBar")
	floatWidth := float64(m.Width)
	barWidth := floatWidth * m.Progress

	intWidth := int(math.Round(barWidth))
	utils.DebugLog("ProgressBar Width: " + strconv.Itoa(m.Width) + " Progress: " + strconv.FormatFloat(m.Progress, 'f', 2, 64) + " intWidth: " + strconv.Itoa(intWidth))
	intEmptyWidth := m.Width - intWidth
	bar := barStyle.Render(strings.Repeat("─", intWidth)) + style.Render(strings.Repeat("─", intEmptyWidth))
	return bar
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if m.Viewport == nil {
		return m, nil
	}
	m.Progress = m.Viewport.ScrollPercent()

	if m.Width == 0 {
		m.Width = utils.TermWidth()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	}
	return m, nil
}
