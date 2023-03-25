package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {

	InitDB()
	var err error
	db, err = sql.Open("sqlite3", "./db.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/board", getBoard).Methods("GET")

	router.HandleFunc("/compose", getCompose).Methods("GET")
	router.HandleFunc("/compose", postCompose).Methods("POST")

	router.HandleFunc("/profile", getProfile).Methods("GET")

	router.HandleFunc("/login", getLogin).Methods("GET")
	router.HandleFunc("/login", postLogin).Methods("POST")

	router.HandleFunc("/register", getRegister).Methods("GET")
	router.HandleFunc("/register", postRegister).Methods("POST")

	router.HandleFunc("/verify/{token}", getVerify).Methods("GET")

	router.PathPrefix("/public/").HandlerFunc(serveStatic)

	fmt.Println("starting on http://localhost:8080")
	http.ListenAndServe(":8080", router)
	select {}
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

func getProfile(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/html/profile.html"))
	err := tmpl.ExecuteTemplate(w, "profile.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/login.html"))
	err := tmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Println(email, password)

	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		http.Error(w, "Incorrect password", http.StatusInternalServerError)
		return
	}

	// Check verified
	fmt.Println("succeeded!")
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/templates/header.tmpl", "public/html/register.html"))
	err := tmpl.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password.", http.StatusInternalServerError)
		return
	}

	token := generateToken()

	stmt, err := db.Prepare("INSERT INTO users(email, password, verification_token) VALUES(?, ?, ?)")
	if err != nil {
		http.Error(w, "Error preparing statement", http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(email, string(hashedPassword), token)
	if err != nil {
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	fmt.Println(email, string(hashedPassword), token)

	// Send verify email.
	sendVerificationEmail(email, token)

	// Check verified
	fmt.Fprintf(w, "User %s registered successfully.", email)
}

func getVerify(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	var email string
	err := db.QueryRow("SELECT email FROM users WHERE verification_token = ?", token).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid token", http.StatusBadRequest)
		} else {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
		}
		return
	}

	stmt, err := db.Prepare("UPDATE users SET verified = 1, verification_token = '' WHERE email = ?")
	if err != nil {
		http.Error(w, "Error preparing statement", http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(email)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
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
