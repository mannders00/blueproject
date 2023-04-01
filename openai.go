package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ProjectPlan struct {
	ImageURLS []struct {
		URL string `json:"url"`
	} `json:"image_urls"`
	Plan PlanResponse `json:"project"`
}

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

type PlanResponse struct {
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

	projectPlan := ProjectPlan{
		ImageURLS: imageResponse.Data,
		Plan:      *planResponse,
	}

	fmt.Println(projectPlan)

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
	Ensure that the output conforms to the following JSON schema (without escape characters in the response):
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

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	choices, ok := response["choices"].([]interface{})
	if !ok {
		fmt.Println("Failure extracting content from OpenAI response.")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		fmt.Println("Failure extracting content from OpenAI response.")
	}

	msg, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		fmt.Println("Failure extracting content from OpenAI response.")
	}

	content, ok := msg["content"]
	if !ok {
		fmt.Println("Failure extracting content from OpenAI response.")
	}

	var plan PlanResponse
	err = json.Unmarshal([]byte(content.(string)), &plan)
	if err != nil {
		return nil, err
	}

	return &plan, nil
}
