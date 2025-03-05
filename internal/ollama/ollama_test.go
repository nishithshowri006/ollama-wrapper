package ollama

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	testModelName = "llama3.2:latest" // Model to use for testing
	testBaseURL   = "http://localhost:11434/api"
)

// TestNewClient tests client initialization
func TestNewClient(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	client := NewClient("", testBaseURL)
	if client == nil {
		t.Fatal("Failed to create Ollama client")
	}

	// Check that initialization with model name also works
	client = NewClient(testModelName, testBaseURL)
	if client == nil {
		t.Fatal("Failed to create Ollama client with model name")
	}
	if client.ModelName != testModelName {
		t.Errorf("Expected model name %s, got %s", testModelName, client.ModelName)
	}
}

// TestListModels tests listing available models
func TestListModels(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	client := NewClient("", testBaseURL)
	models, err := client.ListModels()
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	// Just check that we got some models back
	t.Logf("Found %d models", len(models))
	for i, model := range models {
		if i < 3 { // Just log a few models to avoid flooding output
			t.Logf("Model: %s, Size: %d bytes", model.Name, model.Size)
		}
	}
}

// TestModelExists tests checking if a model exists
func TestModelExists(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	// First get a list of models that are available
	client := NewClient("", testBaseURL)
	models, err := client.ListModels()
	if err != nil || len(models) == 0 {
		t.Skip("No models available to test with")
	}

	// Test with a model that should exist
	existingModel := models[0].Name
	client.ModelName = existingModel
	if !client.ModelExists() {
		t.Errorf("Model %s should exist but reported as not existing", existingModel)
	}

	// Test with a model that shouldn't exist
	client.ModelName = "this_model_definitely_doesnt_exist_12345"
	if client.ModelExists() {
		t.Errorf("Non-existent model was reported as existing")
	}
}

// TestPull tests pulling a model
func TestPull(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	client := NewClient("", testBaseURL)
	err := client.Pull(testModelName)
	if err != nil {
		t.Fatalf("Failed to pull model %s: %v", testModelName, err)
	}

	// Verify the model exists after pulling
	client.ModelName = testModelName
	if !client.ModelExists() {
		t.Errorf("Model %s should exist after pulling but was not found", testModelName)
	}
}

// TestSendMessage tests sending a message and getting a response
func TestSendMessage(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	client := NewClient(testModelName, testBaseURL)

	// Create a simple conversation
	history := []ChatMessage{
		{
			Role:    "user",
			Content: "Hello, can you give me a very brief greeting?",
		},
	}

	response, err := client.SendMessage(history)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if response.Message.Content == "" {
		t.Error("Received empty response content")
	} else {
		t.Logf("Response: %s", response.Message.Content)
	}
}

// TestSendMessageStream tests the streaming functionality
func TestSendMessageStream(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	client := NewClient(testModelName, testBaseURL)

	// Create a simple conversation
	history := []ChatMessage{
		{
			Role:    "user",
			Content: "Hello, count from 1 to 5 very briefly.",
		},
	}

	// Redirect stdout to capture the output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	response, err := client.SendMessageStream(history)

	// Restore stdout and read the captured output
	w.Close()
	os.Stdout = oldStdout
	outBytes, _ := io.ReadAll(r)
	output := string(outBytes)

	if err != nil {
		t.Fatalf("Failed to send message stream: %v", err)
	}

	if response.Message.Content == "" {
		t.Error("Received empty response content")
	}

	t.Logf("Stream captured: %s", output)
	t.Logf("Final response: %s", response.Message.Content)
}

// TestSendMessageStreamReader tests the reader-based streaming functionality
func TestSendMessageStreamReader(t *testing.T) {
	// Skip if Ollama service is not running
	if !isOllamaServiceRunning() {
		t.Skip("Ollama service is not running")
	}

	client := NewClient(testModelName, testBaseURL)

	// Create a simple conversation
	history := []ChatMessage{
		{
			Role:    "user",
			Content: "Say hello in 3 different languages, very briefly.",
		},
	}

	reader, err := client.SendMessageStreamReader(history)
	if err != nil {
		t.Fatalf("Failed to get message stream reader: %v", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	var messageBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()
		var response CompletionResponse

		if err := json.Unmarshal(line, &response); err != nil {
			t.Logf("Failed to decode line: %v", err)
			continue
		}

		messageBuilder.WriteString(response.Message.Content)

		if response.Done {
			break
		}
	}

	completeMessage := messageBuilder.String()
	if completeMessage == "" {
		t.Error("Received empty response from stream reader")
	} else {
		t.Logf("Received response: %s", completeMessage)
	}
}

// Helper function to check if Ollama service is running
func isOllamaServiceRunning() bool {
	_, err := http.Get(testBaseURL + "/tags")
	return err == nil
}
