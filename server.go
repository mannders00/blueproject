package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func generateImageFromPrompt(prompt string, n int) error {

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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

func generateDetailsFromPrompt(problem string, target string, features string, resources string, success string) error {

	fmt.Println("starting generation of project details..")

	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/chat/completions"

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
	`, problem, target, features, resources, success, project_schema)

	payload, err := json.Marshal(Request{
		Model:  "gpt-3.5-turbo",
		Stream: true,
		Messages: []Message{
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

func getIndexHandler(c echo.Context) error {

	// Parse the index.html template file
	tmpl, err := template.ParseFiles("templates/board.html")
	if err != nil {
		http.Error(c.Response().Writer, err.Error(), http.StatusInternalServerError)
	}

	// Execute the template, passing any necessary data as a parameter
	err = tmpl.Execute(c.Response().Writer, nil)
	if err != nil {
		http.Error(c.Response().Writer, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

type InputPrompt struct {
	Problem   string `json:"problem"`
	Target    string `json:"target"`
	Features  string `json:"features"`
	Resources string `json:"resources"`
	Success   string `json:"success"`
}

func postComposeHandler(c echo.Context) error {
	values, err := c.FormValues()
	if err != nil {
		return err
	}

	generateDetailsFromPrompt(values.Get("problem"), values.Get("target"), values.Get("features"), values.Get("resources"), values.Get("success"))
	// generateImageFromPrompt(values.Get("description"), 1)

	return c.File("public/html/compose.html")
}

func (s *Server) Start() {

	app := pocketbase.New()

	//app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
	//	e.Router.POST("/compose.html", postComposeHandler)
	//	e.Router.GET("/compose.html", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
	//	return nil
	//})
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.GET("/", func(c echo.Context) error { return c.File("public/html/compose.html") })
		e.Router.GET("/board", func(c echo.Context) error { return c.File("public/html/board.html") })
		e.Router.GET("/compose", func(c echo.Context) error { return c.File("public/html/compose.html") })
		e.Router.GET("*.js", apis.StaticDirectoryHandler(os.DirFS("./public/js"), false))

		e.Router.POST("/compose", postComposeHandler)

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

const project_schema string = `
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "project_summary": {
      "type": "object",
      "properties": {
        "problem": { "type": "string" },
        "target_audience": { "type": "string" },
        "key_features": {
          "type": "array",
          "items": { "type": "string" }
        },
        "required_support": {
          "type": "array",
          "items": { "type": "string" }
        },
        "success_indicators": {
          "type": "array",
          "items": { "type": "string" }
        }
      },
      "required": ["problem", "target_audience", "key_features", "required_support", "success_indicators"]
    },
    "plan": {
      "type": "object",
      "properties": {
        "goals": {
          "type": "object",
          "properties": {
            "short_term": { "type": "string" },
            "long_term": { "type": "string" }
          },
          "required": ["short_term", "long_term"]
        },
        "tasks": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "name": { "type": "string" },
              "description": { "type": "string" },
              "duration": { "type": "string" }
            },
            "required": ["name", "description", "duration"]
          }
        }
      },
      "required": ["goals", "tasks"]
    }
  },
  "required": ["project_summary", "plan"]
}
`
