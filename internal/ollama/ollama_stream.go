package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// prints completions messages to stand	output
func (o *Ollama) SendMessageStream(history []ChatMessage) (CompletionResponse, error) {
	completionRequest := CompletionRequest{
		Model:    o.ModelName,
		Messages: history,
		Stream:   true,
	}
	body, err := json.Marshal(completionRequest)
	//write file
	var completionResponse CompletionResponse
	if err != nil {
		return completionResponse, err
	}

	req, err := http.NewRequest("POST", o.BaseUrl+"/chat", bytes.NewReader(body))
	if err != nil {
		return completionResponse, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return completionResponse, err
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	var message strings.Builder
	for scanner.Scan() {
		line := scanner.Bytes()
		var temp CompletionResponse
		if err := json.Unmarshal(line, &temp); err != nil {
			log.Println("Failed to decode line")
			return completionResponse, nil
		}
		message.WriteString(temp.Message.Content)
		fmt.Printf("%s", temp.Message.Content)
		if temp.Done {
			completionResponse = temp
			completionResponse.Message.Content = message.String()
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return completionResponse, nil
	}
	fmt.Printf("\n")
	return completionResponse, nil
}
