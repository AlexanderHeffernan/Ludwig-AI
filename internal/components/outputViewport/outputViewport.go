package outputViewport

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"ludwig/internal/components/progressBar"
	"ludwig/internal/types/task"
	"ludwig/internal/utils"
	"ludwig/internal/orchestrator"

	"time"
	"strings"
)

var LOADING_STYLE = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))
var BUBBLE_STYLE = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Width(utils.TermWidth() - 5).
	Height(utils.TermHeight() - 8).
	BorderForeground(lipgloss.Color("62")).
	Padding(0, 1).
	Margin(1, 1)

const VIEWPORT_CONTROLS = "\n(Press Ctrl+S to scroll down, Ctrl+W to scroll up, Esc to exit view)"

type Model struct {
	viewport viewport.Model
	progressBar progressBar.Model
	filePath string
	ViewingTask *task.Task
	fileChangeInfo *utils.FileChangeInfo
	spinner  spinner.Model
}

func NewModel() Model {
	vp := viewport.New(utils.TermWidth()-6, utils.TermHeight()-6)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3
	vp.Style.Padding(0, 0)
	vp.Style.Margin(0, 0)
	//vp.SetContent(utils.OutputLines(strings.Split(utils.ReadFileAsString(filePath), "\n")))
	vp.GotoBottom()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = LOADING_STYLE

	return Model{
		viewport: vp,
		progressBar: progressBar.NewModel(&vp),
		spinner: sp,
	}
}

func (m *Model) SetViewingTask(t *task.Task, filePath string) *Model {
	m.ViewingTask = t
	m.filePath = filePath
	content := utils.OutputLines(strings.Split(utils.ReadFileAsString(filePath), "\n"))
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
	m.fileChangeInfo, _ = utils.InitFileChangeInfo(filePath)
	return m
}

func (m *Model) View() string {
	var s strings.Builder

	s.WriteString(m.progressBar.View())
	// Render full screen output view

	spinnerOn := m.ViewingTask.Status == task.InProgress && orchestrator.IsRunning()

	insideBubble := strings.Builder{}
	insideBubble.WriteString(m.viewport.View())
	if spinnerOn {
		insideBubble.WriteString("\n" + m.spinner.View() + LOADING_STYLE.Render(" Working on it"))
	}

	s.WriteString(BUBBLE_STYLE.Render(insideBubble.String()))
	s.WriteString(VIEWPORT_CONTROLS)
	return s.String()
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmds []tea.Cmd
	var viewportUpdated bool

	//m.progressBar.Progress = m.viewport.ScrollPercent()
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 14
		m.viewport.Height = msg.Height - 6
		viewportUpdated = true
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlS:
			m.viewport.ScrollDown((utils.TermHeight() - 6)/2)
			viewportUpdated = true
		case tea.KeyCtrlW:
			m.viewport.ScrollUp((utils.TermHeight() - 6)/2)
			viewportUpdated = true
		case tea.KeyCtrlC, tea.KeyEsc:
			//m.viewport = &viewport.Model{}
			m.viewport.SetContent("")
			m.fileChangeInfo = nil  // Clean up file change detection
			return m, nil
		}
	case tea.MouseMsg:
		var mouseCmd tea.Cmd
		m.viewport, mouseCmd = m.viewport.Update(msg)
		m.progressBar.Update(msg)
		m.progressBar.Progress = m.viewport.ScrollPercent()
		viewportUpdated = true
		if mouseCmd != nil {
			cmds = append(cmds, mouseCmd)
		}
	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		if spinnerCmd != nil {
			cmds = append(cmds, spinnerCmd)
		}
	}
	
	// Update progress bar after any viewport changes
	//m.progressBar.Update(msg)
	if viewportUpdated {
		// Force progress bar to update its progress from the viewport
	}
	
	// Return all commands batched together
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	// Always start spinner - it will only show when needed based on task status
	m.spinner.Tick()
	return m.spinner.Tick
}

func UpdateViewportWidth(m *Model) {
	termWidth := utils.TermWidth()
	termHeight := utils.TermHeight()
	m.viewport.Width = termWidth - 14
	m.viewport.Height = termHeight - 6
}

func (m *Model) ViewportUpdateLoop()  {
	time.AfterFunc(2*time.Second, func() {
		if m.viewport.Height == 0 || m.fileChangeInfo == nil {
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
		atBottom := scrollPrcnt > 0.95
		content := utils.OutputLines(strings.Split(fileContent, "\n"))
		m.viewport.SetContent(content)
		if atBottom {
			m.viewport.GotoBottom()
		}
		m.ViewportUpdateLoop()
	})
}
