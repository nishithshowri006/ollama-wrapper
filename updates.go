package main

import (
	"fmt"
	"log"
	"strings"
	"terminal-ui/internal/ollama"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

func (m *TerminalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vp   tea.Cmd
		ta   tea.Cmd
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			if !m.TextInput.Focused() {

				return m, tea.Quit
			}
		case tea.KeyEnter.String():
			if m.TextInput.Focused() {
				m.Message = ""
				cm := ollama.ChatMessage{Role: "user", Content: m.TextInput.Value()}
				m.FinalMessage += fmt.Sprintf("\n%s%s\n", userstyle.Render("User: "), strings.TrimSpace(cm.Content))
				m.Viewport.SetContent(m.FinalMessage)
				m.Viewport.GotoBottom()
				m.History = append(m.History, cm)
				cmd = sendMessage(m.History, m.s)
				m.TextInput.Blur()
				m.TextInput.Reset()
				m.SpStatus = spinnerOn
				return m, tea.Batch(cmd, m.Spinner.Tick)
			}
		case tea.KeyEsc.String():
			if m.TextInput.Focused() {
				m.TextInput.Blur()
			} else {
				cmd = m.TextInput.Focus()
			}
			m.Viewport.GotoBottom()
		default:
			if m.TextInput.Focused() {
				m.TextInput, ta = m.TextInput.Update(msg)
				return m, ta
			}
			m.Viewport, vp = m.Viewport.Update(msg)
			return m, tea.Batch(cmd, ta, vp)
		}
	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case reciever:
		m.Chunk = msg.val
		m.Message += m.Chunk
		content := fmt.Sprintf("%s\n%s%s", m.FinalMessage, assistantstyle.Render("Assistant: "), m.Message)
		// m.Viewport.SetContent(m.FinalMessage + "\n" + "Assistant:\t" + strings.TrimSpace(m.Message))
		m.Viewport.SetContent(content)
		m.Viewport.GotoBottom()
		return m, tea.Batch(cmd, listenActivity(m.s))

	case ollama.CompletionResponse:
		m.History = append(m.History, ollama.ChatMessage{Role: msg.Message.Role, Content: msg.Message.Content})
		renderer, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(m.Viewport.Width-10))
		if err != nil {
			log.Fatal(err)
		}
		content, err := renderer.Render(msg.Message.Content)
		m.FinalMessage += fmt.Sprintf("\n%s%s\n", assistantstyle.Render("Assistant: "), strings.TrimSpace(content))
		m.Viewport.SetContent(m.FinalMessage)
		m.Viewport.GotoBottom()
		m.SpStatus = spinnerOff
		return m, tea.Batch(cmd, m.TextInput.Focus(), textinput.Blink)

	case tea.WindowSizeMsg:
		m.Viewport.Height = msg.Height - viewportStyle.GetBorderTopSize() - viewportStyle.GetBorderBottomSize() - m.TextInput.TextStyle.GetHeight()
		m.Viewport.Width = msg.Width - viewportStyle.GetBorderLeftSize() - viewportStyle.GetBorderRightSize()
		m.Viewport.Style = viewportStyle
	}
	m.TextInput, ta = m.TextInput.Update(msg)
	m.Viewport, vp = m.Viewport.Update(msg)
	cmds = append(cmds, cmd, ta, vp)
	return m, tea.Batch(cmds...)
}
