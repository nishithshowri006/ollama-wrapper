package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (o *Ollama) PullStreamReader(Options Opts) (io.ReadCloser, error) {
	if !o.ModelExists() {
		return nil, nil
	}
	if o.ModelName == "" && Options.ModelName != "" {
		o.ModelName = Options.ModelName
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
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama Pull: status code:%d\n", res.StatusCode)
	}
	return res.Body, nil
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
