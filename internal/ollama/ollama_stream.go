package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func (o *Ollama) PullStream(modelName string) (*bytes.Buffer, error) {
	if !o.ModelExists() {
		return nil, nil
	}
	if o.ModelName == "" && modelName != "" {
		o.ModelName = modelName
	} else {
		return nil, fmt.Errorf("Ollama Pull: Haven't Provided Model Name")
	}

	body, err := json.Marshal(struct {
		ModelName string `json:"model"`
	}{
		ModelName: o.ModelName,
	},
	)
	if err != nil {
		return nil, fmt.Errorf("Ollama Pull: Marshal failed, %w\n", err)
	}
	res, err := http.Post(o.BaseUrl+"/pull", "", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Ollama Pull: %w\n", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama Pull: status code:%d\n", res.StatusCode)
	}
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		var pullInfo PullResponse
		if err := json.Unmarshal([]byte(scanner.Text()), &pullInfo); err != nil {
			return nil, fmt.Errorf("Ollama Pull: %w", err)
		}
		fmt.Printf("\r%v", pullInfo.Status)
	}
	return nil, nil
}

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

func (o *Ollama) SendMessageStreamReader(history []ChatMessage) (io.ReadCloser, error) {

	completionRequest := CompletionRequest{

		Model:    o.ModelName,
		Messages: history,
		Stream:   true,
	}
	body, err := json.Marshal(completionRequest)
	//write file
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", o.BaseUrl+"/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
