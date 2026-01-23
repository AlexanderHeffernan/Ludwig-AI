package orchestratorIndicator

import (
	"ludwig/internal/orchestrator"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type animationTickMsg struct{}
type Model struct {
	animationFrame int
}

var frames = [6]string{
	"  ◯ Ludwig composing",
	"  ◎ Ludwig composing",
	"  ◉ Ludwig composing",
	"  ◉ Ludwig composing",
	"  ◎ Ludwig composing",
	"  ◯ Ludwig composing",
}
const frameInterval = 180 * time.Millisecond

var indicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#99ee99")).Bold(true)

func NewModel() *Model {
	return &Model{
		animationFrame: 0,
	}
}

func (m *Model) Init() tea.Cmd {
	// Start the animation loop
	return tea.Every(frameInterval, func(_ time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg.(type) {
	case animationTickMsg:
		m.animationFrame++
		// Always schedule the next tick, but only show animation when running
		return m, tea.Tick(frameInterval, func(time.Time) tea.Msg {
			return animationTickMsg{}
		})

	// You can also handle other messages (keys, window size, etc.)
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	if !orchestrator.IsRunning() {
		return ""
	}
	return indicatorStyle.Render(frames[m.animationFrame%len(frames)])
}
