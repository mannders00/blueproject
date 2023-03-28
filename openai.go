package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ProjectPlan struct {
	ImageURLS []struct {
		URL string `json:"url"`
	} `json:"image_urls"`
	Project string `json:"project"`
}

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

type PlanResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Project struct {
	ProjectSummary ProjectSummary `json:"project_summary"`
	Plan           Plan           `json:"plan"`
}
type ProjectSummary struct {
	Problem           string   `json:"problem"`
	TargetAudience    string   `json:"target_audience"`
	KeyFeatures       []string `json:"key_features"`
	RequiredSupport   []string `json:"required_support"`
	SuccessIndicators []string `json:"success_indicators"`
}
type Plan struct {
	Goals Goals  `json:"goals"`
	Tasks []Task `json:"tasks"`
}
type Goals struct {
	ShortTerm string `json:"short_term"`
	LongTerm  string `json:"long_term"`
}
type Task struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    string `json:"duration"`
}

func GenerateProjectPlan(problem string, target string, features string, success string) (*ProjectPlan, error) {
	imageResponse, err := generateImageFromPrompt(problem, 3)
	if err != nil {
		return nil, err
	}

	planResponse, err := generateDetailsFromPrompt(problem, target, features, success)
	if err != nil {
		return nil, err
	}

	projectPlan := ProjectPlan{
		ImageURLS: imageResponse.Data,
		Project:   planResponse.Choices[0].Message.Content,
	}

	return &projectPlan, nil

}

func generateImageFromPrompt(prompt string, n int) (*ImageResponse, error) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/images/generations"
	payload := []byte(fmt.Sprintf(`{"prompt":"A photorealistic depiction of the following: %s", "n": %d, "size":"512x512"}`, prompt, n))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response ImageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func generateDetailsFromPrompt(problem string, target string, features string, success string) (*PlanResponse, error) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/chat/completions"

	schema, err := ioutil.ReadFile("schema.json")
	if err != nil {
		return nil, err
	}
	schemaString := string(schema)

	message := fmt.Sprintf(`
	{
		"problem": %s
		"target": %s
		"features": %s
		"string": %s
	}
	Generate a summary of the project with an abstract description, required resources (human, financial, material), and a timeline with tasks and expected completion time, in JSON format.
	Ensure that the output conforms to the following JSON schema:
	%s
	`, problem, target, features, success, schemaString)

	payload, err := json.Marshal(ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{Role: "user", Content: message},
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var planResponse PlanResponse
	err = json.Unmarshal(body, &planResponse)

	return &planResponse, nil
}
