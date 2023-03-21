package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {

	// Register HTTP endpoints
	http.HandleFunc("/board", boardHandler)
	http.HandleFunc("/compose", composeHandler)

	// Create access to public static files.
	fileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public", fileServer))

	// Initialize Database
	db := InitDB()
	fmt.Println(db)

	// Start Server
	fmt.Println("listening on port 8080...")
	go http.ListenAndServe(":8080", nil)
	select {}

}

// HTTP /board
func boardHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBoard(w, r)
	}
}

// GET /board
func getBoard(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("public/html/board.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ---

// HTTP /compose
func composeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getCompose(w, r)
	case http.MethodPost:
		postCompose(w, r)
	}
}

// GET /compose
func getCompose(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("public/html/compose.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// POST /compose
func postCompose(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	problem := r.FormValue("problem")
	target := r.FormValue("target")
	features := r.FormValue("features")
	resources := r.FormValue("resources")
	success := r.FormValue("success")

	GenerateDetailsFromPrompt(problem, target, features, resources, success)
	getCompose(w, r)
}
