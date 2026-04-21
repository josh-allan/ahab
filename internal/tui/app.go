package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ahab "github.com/josh-allan/ahab/pkg"
)

type appState int

const (
	stateLoading appState = iota
	stateList
	stateActionRunning
	stateError
)

type paneMode int

const (
	modeInfo paneMode = iota
	modePreview
	modeLogs
)

type composeFile struct {
	path    string
	status  string
	ignored bool
}

type filesLoadedMsg struct{ files []composeFile }
type actionDoneMsg struct{ msg string }
type logTickMsg struct{}
type errMsg struct{ err error }

type Model struct {
	state       appState
	files       []composeFile
	cursor      int
	pane        paneMode
	spinner     spinner.Model
	statusMsg   string
	errMsg      string
	width       int
	height      int
	showHelp    bool
	logStreamer *logStreamer
	preview     string
}

func New() Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	return Model{
		state:   stateLoading,
		spinner: sp,
		pane:    modeInfo,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchFiles())
}

func fetchFiles() tea.Cmd {
	return func() tea.Msg {
		infos, err := ahab.FindComposeFilesForTUI()
		if err != nil {
			return errMsg{err}
		}
		var cfs []composeFile
		for _, info := range infos {
			cfs = append(cfs, composeFile{
				path:    info.Path,
				status:  "unknown",
				ignored: info.Ignored,
			})
		}
		return filesLoadedMsg{cfs}
	}
}

func refreshStatuses(files []composeFile) tea.Cmd {
	return func() tea.Msg {
		updated := make([]composeFile, len(files))
		copy(updated, files)
		for i := range updated {
			updated[i].status = ahab.GetComposeStatus(updated[i].path)
		}
		return filesLoadedMsg{files: updated}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case filesLoadedMsg:
		m.files = msg.files
		m.state = stateList
		m.statusMsg = fmt.Sprintf("%d files", len(m.files))
		return m, refreshStatuses(m.files)

	case actionDoneMsg:
		m.statusMsg = msg.msg
		m.state = stateList
		return m, refreshStatuses(m.files)

	case logTickMsg:
		if m.pane == modeLogs {
			return m, m.logTickCmd()
		}

	case errMsg:
		m.errMsg = msg.err.Error()
		m.state = stateError

	case tea.KeyMsg:
		switch m.state {
		case stateLoading:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case stateList, stateActionRunning:
			return m.updateList(msg)
		case stateError:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "r":
				m.state = stateLoading
				m.errMsg = ""
				return m, tea.Batch(m.spinner.Tick, fetchFiles())
			}
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		switch msg.String() {
		case "?", "esc", "q":
			m.showHelp = false
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		m.stopLogStreamer()
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.preview = ""
			if m.pane == modeLogs {
				m.restartLogStreamer()
			}
		}
	case "down", "j":
		if m.cursor < len(m.files)-1 {
			m.cursor++
			m.preview = ""
			if m.pane == modeLogs {
				m.restartLogStreamer()
			}
		}
	case "tab", "2":
		m.pane = modePreview
		m.loadPreview()
	case "3":
		m.pane = modeLogs
		m.restartLogStreamer()
		return m, m.logTickCmd()
	case "1":
		m.stopLogStreamer()
		m.pane = modeInfo
	case "s":
		return m.runAction("start", "up", "-d")
	case "x":
		return m.runAction("stop", "stop")
	case "d":
		return m.runAction("down", "down")
	case "r":
		return m.runAction("restart", "restart")
	case "p":
		return m.runAction("pull", "pull")
	case "l":
		if m.pane == modeLogs {
			m.stopLogStreamer()
			m.pane = modeInfo
		} else {
			m.pane = modeLogs
			m.restartLogStreamer()
			return m, m.logTickCmd()
		}
	case "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m *Model) runAction(action string, args ...string) (tea.Model, tea.Cmd) {
	if len(m.files) == 0 {
		return *m, nil
	}
	if m.files[m.cursor].ignored {
		m.statusMsg = "cannot run actions on ignored files"
		return *m, nil
	}
	m.state = stateActionRunning
	m.statusMsg = fmt.Sprintf("%s %s...", action, filepath.Base(m.files[m.cursor].path))
	file := m.files[m.cursor].path
	return *m, func() tea.Msg {
		ctx := context.Background()
		if err := ahab.ExecCompose(ctx, io.Discard, io.Discard, file, args...); err != nil {
			return errMsg{err}
		}
		return actionDoneMsg{fmt.Sprintf("%s done", action)}
	}
}

func (m *Model) loadPreview() {
	if m.preview != "" || len(m.files) == 0 {
		return
	}
	data, err := os.ReadFile(m.files[m.cursor].path)
	if err != nil {
		m.preview = fmt.Sprintf("error reading file: %v", err)
		return
	}
	m.preview = string(data)
}

func (m *Model) restartLogStreamer() {
	m.stopLogStreamer()
	if len(m.files) == 0 {
		return
	}
	ls := startLogStreamer(m.files[m.cursor].path)
	m.logStreamer = ls
	if err := ls.run(); err != nil {
		m.logStreamer = nil
		m.statusMsg = fmt.Sprintf("log stream error: %v", err)
	}
}

func (m *Model) stopLogStreamer() {
	if m.logStreamer != nil {
		m.logStreamer.stop()
		m.logStreamer = nil
	}
}

func (m Model) logTickCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return logTickMsg{}
	})
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}
	if m.state == stateLoading {
		return fmt.Sprintf("\n  %s loading compose files...\n\n  press q to quit", m.spinner.View())
	}
	if m.state == stateError {
		return errorStyle.Render("error: "+m.errMsg) + "\n" +
			helpStyle.Render("r retry  q quit")
	}

	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth
	contentHeight := m.height - 2

	left := lipgloss.NewStyle().Width(leftWidth).Height(contentHeight).Render(m.renderList(contentHeight))
	right := lipgloss.NewStyle().Width(rightWidth).Height(contentHeight).Render(m.renderRightPane(contentHeight))
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	statusBar := statusStyle.Render(fmt.Sprintf("  %s  |  %s", m.statusMsg, "? help  q quit"))
	view := body + "\n" + statusBar

	if m.showHelp {
		view = m.renderHelpOverlay(view)
	}
	return view
}

func (m Model) renderList(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("ahab") + "\n\n")

	if len(m.files) == 0 {
		b.WriteString(dimStyle.Render("  no compose files found") + "\n")
	} else {
		maxRows := height - 4
		if maxRows < 1 {
			maxRows = 1
		}
		start := 0
		if m.cursor >= maxRows {
			start = m.cursor - maxRows + 1
		}
		end := start + maxRows
		if end > len(m.files) {
			end = len(m.files)
		}

		for i := start; i < end; i++ {
			f := m.files[i]
			indicator := statusIndicator(f.status)
			name := filepath.Base(f.path)
			if f.ignored {
				name += " [ignored]"
			}
			line := fmt.Sprintf("%s %s", indicator, name)
			if i == m.cursor {
				b.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(normalStyle.Render(line) + "\n")
			}
		}
	}
	return b.String()
}

func (m Model) renderRightPane(height int) string {
	switch m.pane {
	case modePreview:
		return m.renderPreview(height)
	case modeLogs:
		return m.renderLogs(height)
	default:
		return m.renderInfo(height)
	}
}

func (m Model) renderInfo(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("info") + "\n\n")
	if len(m.files) == 0 {
		b.WriteString(dimStyle.Render("  no file selected") + "\n")
		return b.String()
	}
	f := m.files[m.cursor]
	b.WriteString(normalStyle.Render(fmt.Sprintf("Path:   %s", f.path)) + "\n")
	b.WriteString(normalStyle.Render(fmt.Sprintf("Status: %s", f.status)) + "\n\n")
	b.WriteString(helpStyle.Render("s start  x stop  d down  r restart  p pull  l logs"))
	return b.String()
}

func (m Model) renderPreview(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("preview") + "\n\n")
	if m.preview == "" {
		b.WriteString(dimStyle.Render("  loading...") + "\n")
		return b.String()
	}
	lines := strings.Split(m.preview, "\n")
	maxLines := height - 3
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for _, line := range lines {
		b.WriteString(normalStyle.Render(line) + "\n")
	}
	return b.String()
}

func (m Model) renderLogs(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("logs") + "\n\n")
	if m.logStreamer == nil {
		b.WriteString(dimStyle.Render("  no log stream") + "\n")
		return b.String()
	}
	lines := m.logStreamer.buffer.get()
	maxLines := height - 3
	start := 0
	if len(lines) > maxLines {
		start = len(lines) - maxLines
	}
	for _, line := range lines[start:] {
		b.WriteString(logStyle.Render(line) + "\n")
	}
	return b.String()
}

func (m Model) renderHelpOverlay(background string) string {
	help := `
  ahab - keyboard shortcuts

  j/k or ↑/↓   navigate files
  tab/1/2/3    switch pane
  s            start
  x            stop
  d            down
  r            restart
  p            pull
  l            toggle logs
  ?            toggle help
  q/ctrl+c     quit

  press ? or esc to close
`
	overlay := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2).
		Render(help)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}

func statusIndicator(status string) string {
	switch status {
	case "running":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("●")
	case "stopped":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("○")
	case "partial":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("◐")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("?")
	}
}

func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
