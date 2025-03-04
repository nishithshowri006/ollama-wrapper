package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Ollama struct {
	ModelName string
	BaseUrl   string
}
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type CompletionResponse struct {
	Model              string      `json:"model"`
	CreatedAt          time.Time   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done,omitempty"`
	TotalDuration      int64       `json:"total_duration,omitempty"`
	LoadDuration       int         `json:"load_duration,omitempty"`
	PromptEvalCount    int         `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int         `json:"prompt_eval_duration,omitempty"`
	EvalCount          int         `json:"eval_count,omitempty"`
	EvalDuration       int64       `json:"eval_duration,omitempty"`
}
type ModelsMetadata struct {
	Name       string `json:"name,omitempty"`
	ModifiedAt string `json:"modified_at,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Details    struct {
		Format            string `json:"format,omitempty"`
		Family            string `json:"family,omitempty"`
		Families          any    `json:"families,omitempty"`
		ParameterSize     string `json:"parameter_size,omitempty"`
		QuantizationLevel string `json:"quantization_level,omitempty"`
	} `json:"details"`
}
type ModelsList struct {
	Models []ModelsMetadata `json:"models"`
}
type PullRequest struct {
	ModelName string `json:"model"`
}
type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int    `json:"total,omitempty"`
	Completed int    `json:"completed,omitempty"`
}

func NewClient(modelName string, baseUrl string) *Ollama {
	metadata := strings.Split(modelName, ":")
	if len(metadata) < 1 {
		modelName += ":latest"
	}
	if baseUrl == "" {
		baseUrl = "http://localhost:11434/api"
	}
	o := Ollama{ModelName: modelName, BaseUrl: baseUrl}
	// if o.ModelExists
	if o.ModelName == "" {
		return &o
	}
	err := o.Pull(modelName)
	if err != nil {
		log.Fatal(err)
	}
	return &o
}
func (o *Ollama) ListModels() ([]ModelsMetadata, error) {

	res, err := http.Get(o.BaseUrl + "/tags")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var modelsList ModelsList
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &modelsList)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return modelsList.Models, nil
}

func (o *Ollama) ModelExists() bool {
	//make get request
	modelsList, err := o.ListModels()
	if err != nil {
		return false
	}
	for _, model := range modelsList {
		if strings.Contains(model.Name, o.ModelName) {
			return true
		}
	}
	return false
}

func (o *Ollama) Pull(modelName string) error {
	if modelName != "" {
		o.ModelName = modelName
	} else {
		return fmt.Errorf("Ollama Pull: Haven't Provided Model Name")
	}

	if o.ModelExists() {
		return nil
	}
	body, err := json.Marshal(struct {
		ModelName string `json:"model"`
		Stream    bool   `json:"stream"`
	}{
		ModelName: o.ModelName,
		Stream:    true,
	},
	)
	if err != nil {
		return fmt.Errorf("Ollama Pull: Marshal failed, %w\n", err)
	}
	res, err := http.Post(o.BaseUrl+"/pull", "", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("Ollama Pull: %w\n", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama Pull: status code:%d\n", res.StatusCode)
	}
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		var pullInfo PullResponse
		if err := json.Unmarshal([]byte(scanner.Text()), &pullInfo); err != nil {
			return fmt.Errorf("Ollama Pull: %w", err)
		}
		fmt.Printf("\r%v, %.2f/%.2f GB", pullInfo.Status, float32(pullInfo.Completed)/(1024*1024*1024), float32(pullInfo.Total)/(1024*1024*1024))
	}
	fmt.Printf("\r\n")
	return nil
}

func (o *Ollama) SendMessage(history []ChatMessage) (CompletionResponse, error) {
	var cr CompletionResponse
	if o.ModelName == "" {
		return cr, fmt.Errorf("Didn't Provide Model. Please try pulling the model or set model")
	}
	completionRequest := CompletionRequest{
		Model:    o.ModelName,
		Messages: history,
		Stream:   false,
	}
	body, err := json.Marshal(completionRequest)
	//write file
	if err != nil {
		return cr, err
	}

	fp, err := os.OpenFile("log.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return cr, err
	}
	_, err = fp.Write(body)

	if err != nil {
		return cr, err
	}
	req, err := http.NewRequest("POST", o.BaseUrl+"/chat", bytes.NewReader(body))
	if err != nil {
		return cr, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return cr, err
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return cr, err
	}
	_, err = fp.Write(body)

	if err != nil {
		return cr, err
	}
	err = json.Unmarshal(body, &cr)
	if err != nil {
		return cr, err
	}
	return cr, nil
}
