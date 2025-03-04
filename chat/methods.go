package chat

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
)

func (m *TerminalModel) sendMessage() tea.Cmd {
	return func() tea.Msg {
		body, err := client.SendMessageStreamReader(m.History)
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
			m.Message += temp.Message.Content
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
