package main

//
// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"
//
// 	"github.com/charmbracelet/bubbles/spinner"
// 	"github.com/charmbracelet/bubbles/textinput"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"
// 	"github.com/muesli/reflow/wordwrap"
// )
//
// type model struct {
// 	// view_port    viewport.Model
// 	message      chan string
// 	text_input   textinput.Model
// 	sender_style lipgloss.Style
// 	chat_model   Ollama
// 	history      []ChatMessage
// 	spinner      spinner.Model
// 	responses    string
// 	err          error
// }
//
// type Stream struct {
// 	message string
// }
//
// func listenforactivity(sub chan string) tea.Cmd {
// 	return func() tea.Msg {
// 		for i := 0; ; i++ {
// 			time.Sleep(time.Millisecond * 1)
// 			sub <- fmt.Sprintln("Hello World", i)
// 		}
// 	}
// }
//
// func waitForActivity(sub chan string) tea.Cmd {
// 	return func() tea.Msg {
// 		return <-sub
// 	}
// }
//
// func initialModel() model {
// 	ti := textinput.New()
// 	ti.Placeholder = "Start Typing..."
// 	ti.Focus()
// 	ti.CharLimit = 512
// 	ti.Width = 20
// 	h := make([]ChatMessage, 1)
// 	h[0] = ChatMessage{Role: "system", Content: "You are a witty, sarcastic, funny assistant."}
// 	cm := Ollama{model_name: "llama3.2", exists: false, streaming: false}
// 	return model{
// 		text_input: ti, sender_style: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
// 		chat_model: cm,
// 		err:        nil,
// 		history:    h,
// 		spinner:    spinner.New(),
// 		message:    make(chan string),
// 		responses:  "",
// 	}
// }
//
// func (m model) Init() tea.Cmd {
// 	return tea.Batch(
// 		textinput.Blink,
// 		listenforactivity(m.message),
// 		waitForActivity(m.message),
// 	)
// }
//
// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		m.text_input.Width = msg.Width
//
// 	case tea.KeyMsg:
// 		switch msg.Type {
// 		case tea.KeyCtrlC, tea.KeyEsc:
// 			return m, tea.Quit
// 		case tea.KeyEnter:
// 			m.history = append(m.history, ChatMessage{Role: "user", Content: m.text_input.Value()})
// 			m.text_input.Reset()
// 			resp, err := m.chat_model.SendMessage(m.history)
// 			if err != nil {
// 				log.Fatal(err)
// 				os.Exit(1)
// 			}
//
// 			// content := resp.Message.Content
// 			m.history = append(m.history, resp.Message)
// 		}
//
// 		m.text_input, cmd = m.text_input.Update(msg)
// 		return m, tea.Batch(cmd)
// 	}
//
// 	m.text_input, cmd = m.text_input.Update(msg)
// 	return m, cmd
// }
//
// func (m model) View() string {
// 	var view = ""
// 	for _, c := range m.history {
// 		if c.Role == "user" {
// 			view += fmt.Sprint(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).SetString("You"+":"), " ", c.Content, "\n")
// 		} else if c.Role == "assistant" {
//
// 			view += fmt.Sprint(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")).SetString("Assistant:"), " ", c.Content, "\n")
// 		}
// 	}
// 	return wordwrap.String(view+"\n"+m.sender_style.Render(m.text_input.View()), m.text_input.Width-10)
// }
