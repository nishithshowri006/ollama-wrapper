package ollama

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Write([]byte(`{"models":[{"name":"llama3:8b","modified_at":"2023-01-01T00:00:00Z","size":4200000000,"digest":"abc123def456","details":{"format":"gguf","family":"llama","parameter_size":"8B","quantization_level":"Q4_0"}}]}`))
			return
		} else if r.URL.Path == "/api/pull" {
			w.Write([]byte(`{"status":"success","digest":"abc123def456","total":4200000000,"completed":4200000000}`))
			return
		}
	}))
	defer server.Close()

	// Test with default model name format
	client := NewClient("llama3", server.URL+"/api")
	if client.ModelName != "llama3:latest" {
		t.Errorf("Expected model name 'llama3:latest', got '%s'", client.ModelName)
	}
	if client.BaseUrl != server.URL+"/api" {
		t.Errorf("Expected base URL '%s', got '%s'", server.URL+"/api", client.BaseUrl)
	}

	// Test with explicit model version
	client = NewClient("llama3:8b", server.URL+"/api")
	if client.ModelName != "llama3:8b" {
		t.Errorf("Expected model name 'llama3:8b', got '%s'", client.ModelName)
	}
}

func TestModelExists(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Write([]byte(`{"models":[{"name":"llama3:8b"}]}`))
		}
	}))
	defer server.Close()

	// Test existing model
	client := &Ollama{ModelName: "llama3:8b", BaseUrl: server.URL + "/api"}
	if !client.ModelExists() {
		t.Error("Expected ModelExists to return true for llama3:8b model")
	}

	// Test non-existing model
	client = &Ollama{ModelName: "non-existing-model", BaseUrl: server.URL + "/api"}
	if client.ModelExists() {
		t.Error("Expected ModelExists to return false for non-existing model")
	}
}

func TestPull(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			// First call: model doesn't exist, second call: model exists
			w.Write([]byte(`{"models":[]}`))
			return
		} else if r.URL.Path == "/api/pull" {
			w.Write([]byte(`{"status":"downloading manifest","digest":"sha256:abc123def456"}
{"status":"downloading","digest":"sha256:abc123def456","total":4200000000,"completed":2100000000}
{"status":"downloading","digest":"sha256:abc123def456","total":4200000000,"completed":4200000000}
{"status":"verifying sha256 digest","digest":"sha256:abc123def456"}
{"status":"success","digest":"sha256:abc123def456"}`))
			return
		}
	}))
	defer server.Close()

	client := &Ollama{BaseUrl: server.URL + "/api"}
	err := client.Pull("llama3:8b")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if client.ModelName != "llama3:8b" {
		t.Errorf("Expected model name 'llama3:8b', got '%s'", client.ModelName)
	}

	// Test error case - no model name provided
	err = client.Pull("")
	if err == nil {
		t.Error("Expected an error when no model name is provided")
	}
}

func TestSendMessage(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Parse request to verify it's correctly formatted
			decoder := json.NewDecoder(r.Body)
			var req CompletionRequest
			err := decoder.Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode request: %v", err)
			}

			if req.Model != "llama3:8b" {
				t.Errorf("Expected model 'llama3:8b', got '%s'", req.Model)
			}

			if len(req.Messages) != 1 || req.Messages[0].Content != "What are the key features of the Llama 3 model?" {
				t.Errorf("Messages not formatted correctly: %v", req.Messages)
			}

			// Respond with a mock completion that resembles llama3 output
			response := CompletionResponse{
				Model:     "llama3:8b",
				CreatedAt: time.Now(),
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Llama 3 is Meta's latest open source large language model. Key features include improved reasoning abilities, enhanced instruction following, reduced hallucinations, and support for longer context windows compared to previous versions. It was trained on a diverse dataset and comes in various sizes including 8B and 70B parameter versions.",
				},
				Done:               true,
				TotalDuration:      1250000000,
				LoadDuration:       350000000,
				PromptEvalCount:    128,
				PromptEvalDuration: 450000000,
				EvalCount:          320,
				EvalDuration:       780000000,
			}
			jsonResponse, _ := json.Marshal(response)
			w.Write(jsonResponse)
		}
	}))
	defer server.Close()

	client := &Ollama{ModelName: "llama3:8b", BaseUrl: server.URL + "/api"}
	history := []ChatMessage{
		{Role: "user", Content: "What are the key features of the Llama 3 model?"},
	}

	resp, err := client.SendMessage(history)
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}

	if !strings.Contains(resp.Message.Content, "Meta's latest open source large language model") {
		t.Errorf("Unexpected response content: %s", resp.Message.Content)
	}

	if !resp.Done {
		t.Error("Expected Done to be true")
	}
}

func TestSendMessageStream(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Parse request to verify it's correctly formatted
			decoder := json.NewDecoder(r.Body)
			var req CompletionRequest
			err := decoder.Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode request: %v", err)
			}

			if !req.Stream {
				t.Error("Stream should be true for streaming request")
			}

			// Send streaming response
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Error("Expected ResponseWriter to be a Flusher")
				return
			}

			responses := []CompletionResponse{
				{
					Message: ChatMessage{
						Role:    "assistant",
						Content: "Llama 3 ",
					},
				},
				{
					Message: ChatMessage{
						Role:    "assistant",
						Content: "is Meta's latest ",
					},
				},
				{
					Message: ChatMessage{
						Role:    "assistant",
						Content: "open source large language model with improved performance.",
					},
					Done: true,
				},
			}

			for _, resp := range responses {
				jsonResponse, _ := json.Marshal(resp)
				w.Write(jsonResponse)
				w.Write([]byte("\n"))
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	client := &Ollama{ModelName: "llama3:8b", BaseUrl: server.URL + "/api"}
	history := []ChatMessage{
		{Role: "user", Content: "Tell me about Llama 3"},
	}

	resp, err := client.SendMessageStream(history)
	if err != nil {
		t.Errorf("SendMessageStream failed: %v", err)
	}

	expectedContent := "Llama 3 is Meta's latest open source large language model with improved performance."
	if resp.Message.Content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, resp.Message.Content)
	}

	if !resp.Done {
		t.Error("Expected Done to be true")
	}
}

func TestSendMessageStreamReader(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Parse request to verify it's correctly formatted
			decoder := json.NewDecoder(r.Body)
			var req CompletionRequest
			err := decoder.Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode request: %v", err)
			}

			if !req.Stream {
				t.Error("Stream should be true for streaming request")
			}

			// Send streaming response
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Error("Expected ResponseWriter to be a Flusher")
				return
			}

			responses := []string{
				`{"message":{"role":"assistant","content":"Llama 3 is an advanced"},"done":false}`,
				`{"message":{"role":"assistant","content":" language model "},"done":false}`,
				`{"message":{"role":"assistant","content":"developed by Meta AI."},"done":true}`,
			}

			for _, resp := range responses {
				w.Write([]byte(resp + "\n"))
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	client := &Ollama{ModelName: "llama3:8b", BaseUrl: server.URL + "/api"}
	history := []ChatMessage{
		{Role: "user", Content: "What is Llama 3?"},
	}

	responseReader, err := client.SendMessageStreamReader(history)
	if err != nil {
		t.Errorf("SendMessageStreamReader failed: %v", err)
	}
	defer responseReader.Close()

	// Read and process the stream
	scanner := bufio.NewScanner(responseReader)
	var fullMessage strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()
		var response CompletionResponse
		if err := json.Unmarshal(line, &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
			continue
		}

		fullMessage.WriteString(response.Message.Content)

		if response.Done {
			break
		}
	}

	if scanner.Err() != nil {
		t.Errorf("Scanner error: %v", scanner.Err())
	}

	expectedMessage := "Llama 3 is an advanced language model developed by Meta AI."
	if fullMessage.String() != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, fullMessage.String())
	}
}

func TestPullStream(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			// Model doesn't exist
			w.Write([]byte(`{"models":[]}`))
			return
		} else if r.URL.Path == "/api/pull" {
			// Send streaming pull response
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Error("Expected ResponseWriter to be a Flusher")
				return
			}

			responses := []string{
				`{"status":"downloading manifest","digest":"sha256:6b8af5f5382d"}`,
				`{"status":"downloading","digest":"sha256:6b8af5f5382d","total":4200000000,"completed":1050000000}`,
				`{"status":"downloading","digest":"sha256:6b8af5f5382d","total":4200000000,"completed":2100000000}`,
				`{"status":"downloading","digest":"sha256:6b8af5f5382d","total":4200000000,"completed":3150000000}`,
				`{"status":"downloading","digest":"sha256:6b8af5f5382d","total":4200000000,"completed":4200000000}`,
				`{"status":"verifying sha256 digest","digest":"sha256:6b8af5f5382d"}`,
				`{"status":"success","digest":"sha256:6b8af5f5382d"}`,
			}

			for _, resp := range responses {
				w.Write([]byte(resp + "\n"))
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	client := &Ollama{ModelName: "", BaseUrl: server.URL + "/api"}
	_, err := client.PullStream("llama3:8b")

	if err != nil {
		t.Errorf("PullStream failed: %v", err)
	}

	if client.ModelName != "llama3:8b" {
		t.Errorf("Expected model name to be set to 'llama3:8b', got '%s'", client.ModelName)
	}

	// Test error case - no model name provided
	_, err = client.PullStream("")
	if err == nil {
		t.Error("Expected an error when no model name is provided")
	}
}
