package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

func (o *Ollama) Exists() {
	//make get request
	res, err := http.Get(BASE_URL + "/tags")
	if err != nil {
		log.Fatal(err)
	}
	model_list := struct {
		Models []Model `json:"models"`
	}{}
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &model_list)
	if err != nil {
		log.Fatal(err)
	}
	for _, model := range model_list.Models {
		if model.Name == o.model_name {
			o.exists = true
			return
		}
	}
	o.exists = false
}

func (o *Ollama) Pull(model_name string) error {
	if o.exists {
		return nil
	}
	if o.model_name == "" && model_name != "" {
		o.model_name = model_name
	} else {
		return errors.New("Haven't Provided Model Name")
	}
	b := struct {
		Model string `json:"model"`
		// Stream bool   `json:"stream"`
	}{Model: o.model_name} // Stream: false

	body, err := json.Marshal(b)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(BASE_URL+"/pull", "", bytes.NewReader(body))
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		foo := make(map[string]interface{})
		if err := json.Unmarshal([]byte(scanner.Text()), &foo); err != nil {
			return err
		}
	}
	o.exists = true
	return nil
}

func (o *Ollama) SendMessage(history []ChatMessage) (CompletionResponse, error) {
	chatcompletion := Completion{Model: o.model_name, Messages: history, Stream: false}
	body, err := json.Marshal(chatcompletion)
	//write file
	cr := CompletionResponse{}
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
	req, err := http.NewRequest("POST", BASE_URL+"/chat", bytes.NewReader(body))
	if err != nil {
		return cr, err
	}
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return cr, err
	}
	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		return cr, err
	}
	_, err = fp.Write(resp_body)

	if err != nil {
		return cr, err
	}
	err = json.Unmarshal(resp_body, &cr)
	if err != nil {
		return cr, err
	}
	return cr, nil
}
