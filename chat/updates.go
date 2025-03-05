package chat

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vp   tea.Cmd
		ta   tea.Cmd
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	if val, ok := msg.(tea.WindowSizeMsg); ok {
		m.setModelList()
		m.ListModel.SetWidth(val.Width)
		m.ListModel.SetHeight(val.Height)
		m.ListModel.Title = "Available LLM List"
		m.InputView.SetHeight(val.Height / 7)
		m.InputView.SetWidth(val.Width)
		m.ViewportModel.Height = val.Height - m.InputView.Height() - viewportStyle.GetBorderTopSize()
		m.ViewportModel.Width = val.Width - viewportStyle.GetWidth()
		// m.ViewportModel.Style = viewportStyle
		return m, tea.Batch(cmd)
	}
	switch m.WhichView {
	case ListView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			} else if msg.String() == tea.KeyEnter.String() {
				m.WhichView = ChatView
				m.Client.ModelName = m.ListModel.SelectedItem().FilterValue()
				m.InputView.Placeholder = fmt.Sprintf("Chat with %s", m.Client.ModelName)
				return m, tea.Batch(cmd, textarea.Blink)
			}
		}
		m.ListModel, cmd = m.ListModel.Update(msg)

		return m, tea.Batch(cmd)
	case ChatView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case tea.KeyCtrlC.String():
				return m, tea.Quit
			case "q":
				if !m.InputView.Focused() || m.SpStatus == 1 {
					m.WhichView = ListView
					m.FinalMessage.Reset()
					m.Message.Reset()
					m.InputView.Reset()
					m.History = make([]ollama.ChatMessage, 0)
					m.ViewportModel.SetContent("")
					m.ViewportModel.GotoTop()
					m.setModelList()
					return m, nil
				}
			case tea.KeyEnter.String():
				if m.InputView.Focused() {
					if m.InputView.Value() == "/back()" {
						m.WhichView = ListView
						m.FinalMessage.Reset()
						m.Message.Reset()
						m.InputView.Reset()
						m.History = make([]ollama.ChatMessage, 0)
						m.ViewportModel.SetContent("")
						m.ViewportModel.GotoTop()
						m.setModelList()
						return m, nil
					}
					if m.InputView.Value() == "/exit()" || m.InputView.Value() == "/exit" {
						return m, tea.Quit
					}
					if m.InputView.Value() == "/clear()" || m.InputView.Value() == "/clear" {

						m.FinalMessage.Reset()
						m.Message.Reset()
						m.InputView.Reset()
						m.ViewportModel.SetContent("")
						m.ViewportModel.GotoTop()
						m.History = make([]ollama.ChatMessage, 0)
						return m, tea.Batch(cmd, textarea.Blink)
					}
					m.Message.Reset()
					cm := ollama.ChatMessage{Role: "user", Content: m.InputView.Value()}
					m.FinalMessage.WriteString(usercontentstyle.Render("\nUser:", cm.Content, "\n"))
					m.ViewportModel.SetContent(m.FinalMessage.String())
					m.History = append(m.History, cm)
					cmd = m.sendMessage()
					m.InputView.Blur()
					m.InputView.Reset()
					m.SpStatus = spinnerOn
					m.ViewportModel.GotoBottom()
					return m, tea.Batch(cmd, m.Spinner.Tick)
				}
			case tea.KeyEsc.String():
				if m.InputView.Focused() {
					m.InputView.Blur()
				} else {
					cmd = m.InputView.Focus()
					m.ViewportModel.GotoBottom()
				}
			case "up", "down", "pgup", "pgdown":
				m.ViewportModel, _ = m.ViewportModel.Update(msg)
				// default:}
			}
			if m.InputView.Focused() {
				m.InputView, ta = m.InputView.Update(msg)
				return m, tea.Batch(cmd, ta, textarea.Blink)
			}
			m.ViewportModel, vp = m.ViewportModel.Update(msg)
			return m, tea.Batch(cmd, ta, vp)
		case spinner.TickMsg:
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd

		case sender:
			content := fmt.Sprintf("%s%s\n", m.FinalMessage.String(), assistantcontentstyle.Render("Assistant:", m.Message.String()))
			m.ViewportModel.SetContent(content)
			m.ViewportModel.GotoBottom()
			return m, tea.Batch(cmd, m.listenActivity(), vp)

		case ollama.CompletionResponse:
			m.History = append(m.History, ollama.ChatMessage{Role: msg.Message.Role, Content: msg.Message.Content})
			renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(m.ViewportModel.Width-20))
			if err != nil {
				log.Fatal(err)
			}
			content, err := renderer.Render(msg.Message.Content)
			m.ViewportModel.GotoTop()
			m.FinalMessage.WriteString(assistantcontentstyle.Render("Assistant:", strings.TrimSpace(content)))
			m.ViewportModel.SetContent(m.FinalMessage.String())
			m.ViewportModel.GotoBottom()
			m.SpStatus = spinnerOff
			return m, tea.Batch(cmd, m.InputView.Focus(), textinput.Blink, vp)
		default:
			m.InputView, ta = m.InputView.Update(msg)

			m.ViewportModel, vp = m.ViewportModel.Update(msg)

			cmds = append(cmds, cmd, ta, vp)
			return m, tea.Batch(cmds...)
		}
	}
	return m, tea.Batch(cmds...)
}
