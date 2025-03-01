package main

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"
	"terminal-ui/internal/ollama"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type TerminalModel struct {
	TextInput textinput.Model
	Message   string
	Chunk     string
	History   []ollama.ChatMessage
	s         chan sender
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
	ti.Focus()
	s := make(chan sender)
	return TerminalModel{TextInput: ti, s: s}
}

func (m TerminalModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, listenActivity(m.s))
}

func (m TerminalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case reciever:
		m.Chunk = msg.val
		m.Message += m.Chunk
		return m, listenActivity(m.s)
	case ollama.CompletionResponse:
		// m.Message = ""
		m.History = append(m.History, ollama.ChatMessage{Role: msg.Message.Role, Content: msg.Message.Content})
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			m.Message = ""
			cm := ollama.ChatMessage{Role: "user", Content: m.TextInput.Value()}
			m.History = append(m.History, cm)
			m.TextInput.Reset()
			cmd = sendMessage(m.History, m.s)
			return m, tea.Batch(cmd)
		}
		// default:
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, tea.Batch(cmd)
	}
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m TerminalModel) View() string {
	const glamourGutter = 2
	m.TextInput.Focus()
	m.TextInput.Cursor.BlinkCmd()
	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		log.Println(err)
	}
	view, err := renderer.Render(m.Message)
	if err != nil {
		view = m.Message
	}
	return view + "\n" + m.TextInput.View()
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
