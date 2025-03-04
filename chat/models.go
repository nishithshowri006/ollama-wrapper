package chat

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
)

var (
	spinnerOff = 0
	spinnerOn  = 1
)

var (
	ScrollOff = 0
	scrollOn  = 1
)

type TerminalModel struct {
	Viewport     viewport.Model
	InputView    textarea.Model
	Spinner      spinner.Model
	Message      string
	FinalMessage string
	History      []ollama.ChatMessage
	SpStatus     int
	viewLoaded   bool
	position     int
	s            chan sender
}

type sender struct {
}

func (m *TerminalModel) listenActivity() tea.Cmd {
	return func() tea.Msg {
		return <-m.s
	}
}

func InitializeModel() *TerminalModel {
	ta := textarea.New()
	ta.SetHeight(0)
	ta.SetWidth(0)
	ta.SetCursor(0)
	ta.ShowLineNumbers = false
	ta.Focus()
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	vp.Style = viewportStyle
	// vp.HighPerformanceRendering = true
	s := make(chan sender)
	sp := spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerstyle))
	return &TerminalModel{Spinner: sp, s: s, Viewport: vp, InputView: ta}
}

func (m *TerminalModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.listenActivity(), tea.EnterAltScreen)
}

var client = ollama.NewClient("llama3.2", "")
