package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {

	e := echo.New()

	e.GET("/board", getBoard)
	e.GET("/compose", getCompose)
	e.POST("/compose", postCompose)
	e.Static("/public", "public")

	// Initialize Database
	db := InitDB()
	fmt.Println(db)

	// Start Server
	e.Logger.Fatal(e.Start(":8080"))

}

// GET /board
func getBoard(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/board.html"))
	err := tmpl.ExecuteTemplate(c.Response().Writer, "board.html", nil)
	if err != nil {
		http.Error(c.Response().Writer, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

// GET /compose
func getCompose(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/compose.html"))
	err := tmpl.ExecuteTemplate(c.Response().Writer, "compose.html", nil)
	if err != nil {
		http.Error(c.Response().Writer, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

// POST /compose
func postCompose(c echo.Context) error {

	problem := c.FormValue("problem")
	target := c.FormValue("target")
	features := c.FormValue("features")
	resources := c.FormValue("resources")
	success := c.FormValue("success")

	//GenerateDetailsFromPrompt(problem, target, features, resources, success)
	fmt.Println(problem, target, features, resources, success)
	getCompose(c)
	return nil
}
