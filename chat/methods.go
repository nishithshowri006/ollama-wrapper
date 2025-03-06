package chat

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
)

func (m *Model) listenActivity() tea.Cmd {
	return func() tea.Msg {
		return <-m.s
	}
}
func (m *Model) sendMessage() tea.Cmd {

	return func() tea.Msg {
		body, err := m.Client.SendMessageStreamReader(m.History)
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
			m.Message.WriteString(temp.Message.Content)
			if temp.Done {
				response = temp
				response.Message.Content = message.String()
				break
			}
			m.s <- sender{}

		}
		return response
	}
}
func (m *Model) setModelList() {
	mlist := m.Client.ModelsList
	items := make([]list.Item, len(mlist))
	for i := range len(mlist) {
		items[i] = ItemModel{mlist[i]}
	}
	m.ListModel.SetItems(items)
}
