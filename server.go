package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	ory "github.com/ory/client-go"
)

var db *sql.DB

type App struct {
	ory *ory.APIClient
}

const proxyPort = 4000

func main() {

	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{{URL: fmt.Sprintf("http://localhost:%d/.ory", proxyPort)}}
	fmt.Println(c.Servers)
	app := &App{
		ory: ory.NewAPIClient(c),
	}

	router := mux.NewRouter()

	router.HandleFunc("/", getIndex).Methods("GET")

	router.HandleFunc("/board", getBoard).Methods("GET")

	router.HandleFunc("/compose", getCompose).Methods("GET")
	router.HandleFunc("/compose", postCompose).Methods("POST")

	router.HandleFunc("/profile", app.sessionMiddleware(app.profileHandler())).Methods("GET")

	router.PathPrefix("/public/").HandlerFunc(serveStatic)

	fmt.Println("starting on http://localhost:8080")
	http.ListenAndServe(":8080", router)
	select {}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/index.html"))
	err := tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getBoard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/board.html"))
	err := tmpl.ExecuteTemplate(w, "board.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getCompose(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/compose.html"))
	err := tmpl.ExecuteTemplate(w, "compose.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postCompose(w http.ResponseWriter, r *http.Request) {

	problem := r.FormValue("problem")
	target := r.FormValue("target")
	features := r.FormValue("features")
	resources := r.FormValue("resources")
	success := r.FormValue("success")

	fmt.Println(problem, target, features, resources, success)
	getCompose(w, r)
}

func (app *App) getProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("public/html/profile.html"))
		session, err := json.Marshal(getSession(r.Context()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = tmpl.ExecuteTemplate(w, "profile.html", string(session))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	filePath := filepath.Join(".", path)

	if strings.HasPrefix(path, "/public/css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasPrefix(path, "/public/js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasPrefix(path, "/public/images") {
		w.Header().Set("Content-Type", "image/jpeg")
	}

	http.ServeFile(w, r, filePath)
}
