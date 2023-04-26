package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ory/client-go"
	ory "github.com/ory/client-go"
)

type App struct {
	db  *sql.DB
	ory *ory.APIClient
}

var app App

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	c := ory.NewConfiguration()
	c.Servers = client.ServerConfigurations{
		{URL: fmt.Sprint("http://localhost:4433")},
	}
	//c.Servers = client.ServerConfigurations{
	//	{URL: fmt.Sprintf("https://%s.projects.oryapis.com", os.Getenv("ORY_PROJECT_SLUG"))},
	//}

	app = App{
		db:  InitDB(),
		ory: ory.NewAPIClient(c),
	}

	router := mux.NewRouter()

	router.HandleFunc("/", getIndex).Methods("GET")

	router.HandleFunc("/board", getBoard).Methods("GET")

	router.HandleFunc("/compose", app.sessionMiddleware(app.getCompose())).Methods("GET")
	router.HandleFunc("/compose", app.sessionMiddleware(app.postCompose())).Methods("POST")
	router.HandleFunc("/generationStatus", getGenerationStatus).Methods("GET")

	router.HandleFunc("/project/{id}", getProject).Methods("GET")

	router.HandleFunc("/profile", app.sessionMiddleware(app.getProfile())).Methods("GET")

	router.PathPrefix("/public/").HandlerFunc(serveStatic)
	router.PathPrefix("/data/").HandlerFunc(serveStatic)

	fmt.Println("starting on http://:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/templates/footer.tmpl", "public/html/index.html"))
	err := tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getBoard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/templates/footer.tmpl", "public/html/board.html"))
	err := tmpl.ExecuteTemplate(w, "board.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) getCompose() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/templates/footer.tmpl", "public/html/compose.html"))
		err := tmpl.ExecuteTemplate(w, "compose.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

var taskStatus = make(map[string]string)

func (app *App) postCompose() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r.Context())
		user_id := session.Identity.Id

		problem := r.FormValue("problem")
		target := r.FormValue("target")
		features := r.FormValue("features")
		success := r.FormValue("success")

		unique_id, err := GenerateRandomString(32)
		// TODO: check if unique_id is unique in database
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		go func() {
			project, err := GenerateProjectPlan(unique_id, problem, target, features, success)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			err = SaveProject(app.db, user_id, unique_id, project)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			taskStatus[unique_id] = "completed"
		}()

		// Return a pending response with the task ID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "pending",
			"task_id": unique_id,
		})
	}
}

func getGenerationStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	status := taskStatus[taskID]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
	})
}

func getProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	project, err := LoadProject(app.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/templates/footer.tmpl", "public/html/project.html"))
	err = tmpl.ExecuteTemplate(w, "project.html", project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ProfilePayload struct {
	Session  *ory.Session
	Projects []Project
}

func (app *App) getProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r.Context())
		if session == nil {
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}
		user_id := session.Identity.Id

		rows, err := app.db.Query("SELECT id, user_id, data FROM projects WHERE user_id = ?", user_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer rows.Close()

		var projects []Project

		for rows.Next() {
			var ID string
			var UserID string
			var PlanStr string
			if err := rows.Scan(&ID, &UserID, &PlanStr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			var Plan ProjectPlan
			err := json.Unmarshal([]byte(PlanStr), &Plan)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			p := Project{
				ID:     ID,
				UserID: UserID,
				Plan:   Plan,
			}

			projects = append(projects, p)
		}

		payload := ProfilePayload{
			Session:  session,
			Projects: projects,
		}

		tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/templates/footer.tmpl", "public/html/profile.html"))

		err = tmpl.ExecuteTemplate(w, "profile.html", payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	} else if strings.HasPrefix(path, "/data/images") {
		w.Header().Set("Content-Type", "image/png")
	}

	http.ServeFile(w, r, filePath)
}
