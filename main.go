package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	// "terminal-ui/cmd"
	"bufio"
	"fmt"
	"time"
)

// TODO:
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Completion struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type CompletionResponse struct {
	Model   string      `json:"model"`
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}
type Ollama struct {
	model_name string
	streaming  bool
	exists     bool
}

type Model struct {
	Name        string    `json:"name"`
	Modified_At time.Time `json:"modified_at"`
	Size        int       `json:"size"`
}

func run_cli() {

	chat := Ollama{model_name: "llama3.2", exists: true}
	messages := make([]ChatMessage, 0)
	for {
		fmt.Printf(">> ")
		scanner := bufio.NewReader(os.Stdin)
		var user_message string
		user_message, err := scanner.ReadString('\n')
		if err != nil {
			panic(err)
		}
		messages = append(messages, ChatMessage{Role: "user", Content: user_message})
		chat_response, err := chat.SendMessage(messages)
		if err != nil {
			panic(err)
		}
		messages = append(messages, chat_response.Message)
		fmt.Printf(">> %s\n", chat_response.Message.Content)
	}
}

var BASE_URL = "http://localhost:11434/api"

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
