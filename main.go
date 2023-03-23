package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	ory "github.com/ory/kratos-client-go"
)

func main() {

	k := NewMiddleware()

	mux := http.NewServeMux()

	mux.HandleFunc("/compose", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCompose(w, r)
		case http.MethodPost:
			postCompose(w, r)
		}
	})
	mux.HandleFunc("/profile", k.sessionMiddleware(getProfile))
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

// GET /profile
func getProfile(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/html/profile.html"))
	err := tmpl.ExecuteTemplate(w, "profile.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

type kratosMiddleware struct {
	ory *ory.APIClient
}

func NewMiddleware() *kratosMiddleware {
	configuration := ory.NewConfiguration()
	configuration.Servers = []ory.ServerConfiguration{
		{
			URL: "http://127.0.0.1:4433", // Kratos Public API
		},
	}
	return &kratosMiddleware{
		ory: ory.NewAPIClient(configuration),
	}
}
func (k *kratosMiddleware) sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := k.validateSession(r)
		if err != nil {
			http.Redirect(w, r, "http://127.0.0.1:8080/login", http.StatusMovedPermanently)
			return
		}
		if !*session.Active {
			http.Redirect(w, r, "http://your_endpoint", http.StatusMovedPermanently)
			return
		}
		next(w, r)
	}
}
func (k *kratosMiddleware) validateSession(r *http.Request) (*ory.Session, error) {
	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		return nil, err
	}
	if cookie == nil {
		return nil, errors.New("no session found in cookie")
	}
	resp, _, err := k.ory.FrontendApi.ToSession(context.Background()).Cookie(cookie.String()).Execute()
	if err != nil {
		return nil, err
	}
	return resp, nil
}
