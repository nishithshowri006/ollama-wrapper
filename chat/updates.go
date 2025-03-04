package chat

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
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
			if !m.InputView.Focused() || m.SpStatus == 1 {

				return m, tea.Quit
			}

		case tea.KeyEnter.String():
			if m.InputView.Focused() {
				m.Message = ""
				cm := ollama.ChatMessage{Role: "user", Content: m.InputView.Value()}
				m.Viewport.GotoTop()
				m.FinalMessage += fmt.Sprintf("\n%s%s\n", userstyle.Render("User: "), strings.TrimSpace(cm.Content))
				m.Viewport.SetContent(m.FinalMessage)
				m.Viewport.GotoBottom()
				// vp = viewport.Sync(m.Viewport)
				m.History = append(m.History, cm)
				cmd = m.sendMessage()
				m.InputView.Blur()
				m.InputView.Reset()
				m.SpStatus = spinnerOn
				return m, tea.Batch(cmd, m.Spinner.Tick)
			}
		case tea.KeyEsc.String():
			if m.InputView.Focused() {
				m.InputView.Blur()
			} else {
				cmd = m.InputView.Focus()
			}
			m.Viewport.GotoBottom()
		default:
			if m.InputView.Focused() {
				m.InputView, ta = m.InputView.Update(msg)
				return m, ta
			}
			m.Viewport, vp = m.Viewport.Update(msg)
			return m, tea.Batch(cmd, ta, vp)
		}
	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case sender:
		content := fmt.Sprintf("%s\n%s%s", m.FinalMessage, assistantstyle.Render("Assistant: "), m.Message)
		m.Viewport.GotoTop()
		m.Viewport.SetContent(content)
		m.Viewport.GotoBottom()
		// vp = viewport.Sync(m.Viewport)
		return m, tea.Batch(cmd, m.listenActivity(), vp)

	case ollama.CompletionResponse:
		m.History = append(m.History, ollama.ChatMessage{Role: msg.Message.Role, Content: msg.Message.Content})
		renderer, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(m.Viewport.Width-10))
		if err != nil {
			log.Fatal(err)
		}
		content, err := renderer.Render(msg.Message.Content)
		m.Viewport.GotoTop()
		m.FinalMessage += fmt.Sprintf("\n%s%s\n", assistantstyle.Render("Assistant: "), strings.TrimSpace(content))
		m.Viewport.SetContent(m.FinalMessage)
		m.Viewport.GotoBottom()
		// vp = viewport.Sync(m.Viewport)
		m.SpStatus = spinnerOff
		return m, tea.Batch(cmd, m.InputView.Focus(), textinput.Blink, vp)

	case tea.WindowSizeMsg:
		m.InputView.SetHeight(msg.Height / 8)
		m.InputView.SetWidth(msg.Width)
		m.Viewport.Height = msg.Height - m.InputView.Height() - viewportStyle.GetBorderBottomSize()
		m.Viewport.Width = msg.Width - viewportStyle.GetWidth()
	}
	m.InputView, ta = m.InputView.Update(msg)

	m.Viewport, vp = m.Viewport.Update(msg)

	cmds = append(cmds, cmd, ta, vp)
	return m, tea.Batch(cmds...)
}
