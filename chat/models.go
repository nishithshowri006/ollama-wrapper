package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type Model struct {
	ViewportModel viewport.Model
	ListModel     list.Model
	MaxHeight     int
	MaxWidth      int
	InputView     textarea.Model
	Spinner       spinner.Model
	Message       strings.Builder
	FinalMessage  strings.Builder
	History       []ollama.ChatMessage
	SpStatus      int
	modelLoaded   bool
	WhichView     int
	position      int
	s             chan sender
	Client        *ollama.Ollama
}

var (
	ListView = 0
	ChatView = 1
)

type ItemModel struct {
	ollama.ModelsMetadata
}

func (i ItemModel) Title() string {
	return i.Name
}

func (i ItemModel) Description() string {
	return fmt.Sprintf("Size: %d\nFamily: %s\nQuantizationLevels: %s\n",
		i.Size, i.Details.Family, i.Details.QuantizationLevel,
	)
}

func (i ItemModel) FilterValue() string {
	return i.Name
}

type sender struct {
}

func (m *Model) listenActivity() tea.Cmd {
	return func() tea.Msg {
		return <-m.s
	}
}

func InitializeModel() *Model {
	var m Model
	ta := textarea.New()
	ta.SetHeight(0)
	ta.SetWidth(0)
	ta.SetCursor(0)
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.FocusedStyle = textarea.Style{Base: textareaStyle, Placeholder: lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(0))}
	ta.BlurredStyle = textarea.Style{Base: textareaBlurStyle.Faint(true), Placeholder: lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(0))}
	client := ollama.NewClient("", "")
	s := make(chan sender)
	sp := spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerstyle))
	vp := viewport.New(0, 0)
	vp.Style = viewportStyle
	m.Spinner = sp
	m.s = s
	m.Client = client
	m.InputView = ta
	m.ViewportModel = vp
	m.ListModel = list.New(nil, list.NewDefaultDelegate(), 0, 0)

	//ollama List settings
	m.WhichView = ListView
	return &m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.listenActivity(), tea.EnterAltScreen)
}
