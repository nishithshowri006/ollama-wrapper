package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"terminal-ui/internal/ollama"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

var (
	spinnerOff = 0
	spinnerOn  = 1
)

var (
	ScrollOff = 0
	scrollOn  = 1
)
var (
	resizeNo  = 0
	resizeYes = 1
)

type TerminalModel struct {
	Viewport     viewport.Model
	InputView    textarea.Model
	TextInput    textinput.Model
	Spinner      spinner.Model
	Message      string
	FinalMessage string
	Chunk        string
	History      []ollama.ChatMessage
	SpStatus     int
	scrollStatus int
	viewResize   int
	s            chan sender
}

type reciever struct {
	val string
}

type sender struct {
	val string
}

func listenActivity(r chan sender) tea.Cmd {
	return func() tea.Msg {
		val := <-r
		return reciever{val: val.val}
	}
}

func initialModel() TerminalModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your input here.."

	ti.Focus()
	s := make(chan sender)
	sp := spinner.New(spinner.WithSpinner(spinner.Line), spinner.WithStyle(spinnerstyle))
	return TerminalModel{TextInput: ti, Spinner: sp, s: s}
}

func (m TerminalModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, listenActivity(m.s), tea.EnterAltScreen)
}

func (m TerminalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.viewResize == resizeYes {
			m.Viewport.Height = msg.Height - viewportStyle.GetBorderTopSize() - viewportStyle.GetBorderBottomSize()
			m.Viewport.Width = msg.Width
		}
		if m.viewResize == resizeNo {

			m.Viewport = viewport.New(msg.Width, msg.Height-viewportStyle.GetBorderBottomSize()-viewportStyle.GetBorderTopSize())
			m.viewResize = resizeYes
		}

		m.Viewport.MouseWheelEnabled = true
		// m.Viewport.HighPerformanceRendering = true
		m.Viewport.Style = viewportStyle
		m.Viewport.Style.Width(msg.Width - 10)
		m.Viewport.GotoBottom()
	case reciever:
		m.Chunk = msg.val
		m.Message += m.Chunk
		m.Viewport.SetContent(m.FinalMessage + "\n" + "Assistant: " + strings.TrimSpace(m.Message))
		m.Viewport.GotoBottom()
		return m, listenActivity(m.s)
	case ollama.CompletionResponse:
		m.History = append(m.History, ollama.ChatMessage{Role: msg.Message.Role, Content: msg.Message.Content})
		m.Spinner = spinner.New(spinner.WithSpinner(spinner.MiniDot))
		renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(m.Viewport.Style.GetWidth()-10))
		if err != nil {
			log.Fatal(err)
		}
		content, err := renderer.Render(msg.Message.Content)
		m.FinalMessage += "\n" + assistantstyle.Render("Assistant:", strings.TrimSpace(content))
		m.Viewport.SetContent(m.FinalMessage)
		m.SpStatus = spinnerOff
		m.Viewport.GotoBottom()
		return m, tea.Batch(cmd, m.TextInput.Focus(), textinput.Blink)

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			if !m.TextInput.Focused() {

				return m, tea.Quit
			}
		case tea.KeyEnter.String():
			m.Message = ""
			cm := ollama.ChatMessage{Role: "user", Content: m.TextInput.Value()}
			m.FinalMessage += fmt.Sprintf("\n%s", userstyle.Render("User:", strings.TrimSpace(cm.Content)))
			m.Viewport.SetContent(fmt.Sprintf("\n%s", userstyle.Render("User:", strings.TrimSpace(cm.Content))))
			m.History = append(m.History, cm)
			cmd = sendMessage(m.History, m.s)
			m.TextInput.Blur()
			m.TextInput.Reset()
			m.SpStatus = 1
			return m, tea.Batch(cmd, m.Spinner.Tick)
		case tea.KeyEsc.String():
			if m.TextInput.Focused() {

				m.TextInput.Blur()
			} else {
				cmd = m.TextInput.Focus()
			}
			return m, cmd
		}
		if msg.String() == tea.KeyUp.String() || msg.String() == "k" && !m.TextInput.Focused() {
			m.Viewport.ViewUp()
			m.Viewport, cmd = m.Viewport.Update(msg)
			return m, cmd
		}
		if msg.String() == tea.KeyDown.String() || msg.String() == "j" && !m.TextInput.Focused() {
			m.Viewport.ViewDown()
			m.Viewport, cmd = m.Viewport.Update(msg)
			return m, cmd
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, tea.Batch(cmd)
	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
	m.TextInput, cmd = m.TextInput.Update(msg)
	cmds = append(cmds, cmd)
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m TerminalModel) View() string {
	if m.Message == "" {
		return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.TextInput.View())
	}

	if m.SpStatus == 0 {
		return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.TextInput.View())
	}
	return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.Spinner.View())
}

var client = ollama.NewClient("llama3.2", "")

func sendMessage(history []ollama.ChatMessage, s chan sender) tea.Cmd {
	return func() tea.Msg {
		body, err := client.SendMessageStreamReader(history)
		defer body.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		scanner := bufio.NewScanner(body)
		var message strings.Builder
		var response ollama.CompletionResponse
		for scanner.Scan() {
			line := scanner.Bytes()
			var temp ollama.CompletionResponse
			if err := json.Unmarshal(line, &temp); err != nil {
				log.Println("Failed to decode line")
				return err
			}
			message.WriteString(temp.Message.Content)
			s <- sender{temp.Message.Content}
			if temp.Done {
				response = temp
				response.Message.Content = message.String()
				break
			}
		}
		return response
	}
}
