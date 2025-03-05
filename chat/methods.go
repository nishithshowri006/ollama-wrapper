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
			m.s <- sender{}
			if temp.Done {
				response = temp
				response.Message.Content = message.String()
				break
			}
		}
		return response
	}
}
func (m *Model) setModelList() {
	mlist, err := m.Client.ListModels()
	if err != nil {
		log.Fatal(err)
	}
	items := make([]list.Item, len(mlist))
	for i := range len(mlist) {
		items[i] = ItemModel{mlist[i]}
	}
	m.ListModel.SetItems(items)
}
