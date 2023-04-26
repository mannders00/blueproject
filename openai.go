package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ProjectPlan struct {
	ImageURLS []struct {
		URL string `json:"url"`
	} `json:"image_urls"`
	Plan PlanResponse `json:"project"`
}

func GenerateProjectPlan(unique_id string, problem string, target string, features string, success string) (*ProjectPlan, error) {

	planResponse, err := generateDetailsFromPrompt(problem, target, features, success)

	imageResponse, err := generateImageFromPrompt(planResponse.ImageDescription, 3, unique_id)
	if err != nil {
		return nil, err
	}

	projectPlan := ProjectPlan{
		ImageURLS: imageResponse.Data,
		Plan:      *planResponse,
	}

	return &projectPlan, nil
}

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func generateImageFromPrompt(prompt string, n int, unique_id string) (*ImageResponse, error) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/images/generations"
	payload := []byte(fmt.Sprintf(`{"prompt":"%s", "n": %d, "size":"512x512"}`, prompt, n))

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

	destFolder := fmt.Sprintf("./data/%s/images", unique_id)
	err = os.MkdirAll(destFolder, 0755)
	if err != nil {
		return nil, err
	}

	imageURLS := response.Data
	for _, imageURL := range imageURLS {
		err := downloadAndSaveImage(imageURL.URL, destFolder)
		if err != nil {
			return nil, err
		}
	}

	filePaths, err := getAllFiles(destFolder)
	for i := 0; i < len(filePaths); i++ {
		response.Data[i].URL = fmt.Sprintf("/data/%s/images/%s", unique_id, filePaths[i])
	}

	return &response, nil
}

func downloadAndSaveImage(imageURL, destFolder string) error {
	// Fetch the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return err
	}

	// Get the image file name from the URL
	fileName := filepath.Join(destFolder, filepath.Base(parsedURL.Path))
	if strings.TrimSpace(filepath.Ext(fileName)) == "" {
		return fmt.Errorf("unable to determine file extension for URL: %s", imageURL)
	}

	// Create the image file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Save the image data to the file
	_, err = io.Copy(file, resp.Body)
	return err
}

func getAllFiles(folderPath string) ([]string, error) {
	var filePaths []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return err
			}
			filePaths = append(filePaths, relativePath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filePaths, nil
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
	ProjectSummary   ProjectSummary `json:"project_summary"`
	Plan             Plan           `json:"plan"`
	ImageDescription string         `json:"image_description"`
}

type ProjectSummary struct {
	Problem           string   `json:"problem"`
	Title             string   `json:"title"`
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
	Generate a summary of the project with an abstract description, title, required resources (human, financial, material), and a timeline with tasks and expected completion time, in JSON format.
	Additionally generate an image_description which creates a DALL-E prompt to create a photorealistic and artistic depiction of the project.
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
