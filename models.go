package main

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"
	"terminal-ui/internal/ollama"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	TextInput    textinput.Model
	Spinner      spinner.Model
	Message      string
	FinalMessage string
	Chunk        string
	History      []ollama.ChatMessage
	SpStatus     int
	viewLoaded   bool
	position     int
	Cursor       cursor.Model
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

func initialModel() *TerminalModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your input here.."
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true

	ti.Focus()
	s := make(chan sender)
	sp := spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerstyle))
	return &TerminalModel{TextInput: ti, Spinner: sp, s: s, Viewport: vp}
}

func (m *TerminalModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, listenActivity(m.s), tea.EnterAltScreen)
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
