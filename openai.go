package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func GenerateImageFromPrompt(prompt string, n int) error {

	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/images/generations"
	payload := []byte(fmt.Sprintf(`{"prompt":"%s", "n": %d, "size":"512x512"}`, prompt, n))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response ImageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	fmt.Println(response.Data)

	return nil
}

func GenerateDetailsFromPrompt(problem string, target string, features string, resources string, success string) error {

	fmt.Println("starting generation of project details..")

	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/chat/completions"

	schema, err := ioutil.ReadFile("schema.json")
	if err != nil {
		return err
	}
	schemaString := string(schema)

	message := fmt.Sprintf(`
	{
		"problem": %s
		"target": %s
		"features": %s
		"resources": %s
		"string": %s
	}
	Generate a summary of the project with an abstract description, required resources (human, financial, material), and a timeline with tasks and expected completion time, in JSON format.
	Ensure that the output conforms to the following JSON schema:
	%s
	`, problem, target, features, resources, success, schemaString)

	payload, err := json.Marshal(ChatRequest{
		Model:  "gpt-3.5-turbo",
		Stream: true,
		Messages: []ChatMessage{
			{Role: "user", Content: message},
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response map[string]interface{}
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return err
	}

	fmt.Println(response)

	return nil
}
