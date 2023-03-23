package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/board", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getBoard(w, r)
		}
	})
	mux.HandleFunc("/compose", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCompose(w, r)
		case http.MethodPost:
			postCompose(w, r)
		}
	})
	mux.HandleFunc("/login", getLogin)
	mux.HandleFunc("/register", getRegister)

	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("/public/", http.StripPrefix("/public", fs))

	// Initialize Database
	db := InitDB()
	fmt.Println(db)

	// Start Servea
	fmt.Println("starting on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

// GET /board
func getBoard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/board.html"))
	err := tmpl.ExecuteTemplate(w, "board.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GET /compose
func getCompose(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/compose.html"))
	err := tmpl.ExecuteTemplate(w, "compose.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// POST /compose
func postCompose(w http.ResponseWriter, r *http.Request) {

	problem := r.FormValue("problem")
	target := r.FormValue("target")
	features := r.FormValue("features")
	resources := r.FormValue("resources")
	success := r.FormValue("success")

	//GenerateDetailsFromPrompt(problem, target, features, resources, success)
	fmt.Println(problem, target, features, resources, success)
	getCompose(w, r)
}

// GET /login
func getLogin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/login.html"))
	err := tmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GET /register
func getRegister(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/register.html"))
	err := tmpl.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
