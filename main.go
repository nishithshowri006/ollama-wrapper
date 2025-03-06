package main

import (
	"bufio"
	"fmt"
	"log"

	// "log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishithshowri006/ollama-wrapper/chat"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
)

func run_cli() {

	client := ollama.NewClient(&ollama.Opts{})
	messages := make([]ollama.ChatMessage, 0)
	for {
		fmt.Printf(">> ")
		scanner := bufio.NewReader(os.Stdin)
		var user_message string
		user_message, err := scanner.ReadString('\n')
		if err != nil {
			panic(err)
		}
		messages = append(messages, ollama.ChatMessage{Role: "user", Content: user_message})
		chat_response, err := client.SendMessageStream(messages)
		if err != nil {
			panic(err)
		}
		messages = append(messages, chat_response.Message)
		// fmt.Printf(">> %s\n", chat_response.Message.Content)
	}
}

var BASE_URL = "http://localhost:11434/api"

func smn() {

	p := tea.NewProgram(chat.InitializeModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
func main() {
	// run_cli()
	smn()
}
