package progressBar

import (
	"strings"
	"math"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"ludwig/internal/utils"
)

type Model struct {
	Progress float64
	Width int
}

func NewModel(v *viewport.Model) Model {
	return Model{
		Progress: v.ScrollPercent(),
		Width: 0,
	}
}

var barStyle = lipgloss.NewStyle().Bold(true)
var style = lipgloss.NewStyle().Faint(true)

func (m *Model) View() string {
	if m.Width == 0 {
		m.Width = utils.TermWidth()
	}
	floatWidth := float64(m.Width)
	//m.Progress = m.Viewport.ScrollPercent()
	barWidth := floatWidth * m.Progress

	intWidth := int(math.Round(barWidth))
	intEmptyWidth := m.Width - intWidth
	bar := barStyle.Render(strings.Repeat("─", intWidth)) + style.Render(strings.Repeat("─", intEmptyWidth))
	return bar
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlW:
			m.Width = utils.TermWidth();
		case tea.KeyCtrlS:
			m.Width = utils.TermWidth();
		}
	}
	return m, nil
}
