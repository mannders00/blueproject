package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db sql.DB

type Contractor struct {
	ID          int
	Name        string
	Description string
	Lng         float64
	Lat         float64
}

func getIndexHandler(w http.ResponseWriter, r *http.Request) {

	// Open database
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create initial tabe
	query := `
		CREATE TABLE IF NOT EXISTS contractors
		(id INTEGER PRIMARY KEY, name TEXT, description TEXT, lng NUMERIC, lat NUMERIC)
	`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the index.html template file
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, name, description, lng, lat FROM contractors")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var ctrs []Contractor
	for rows.Next() {
		var c Contractor
		err = rows.Scan(&c.ID, &c.Name, &c.Description, &c.Lng, &c.Lat)
		if err != nil {
			log.Fatal(err)
		}
		ctrs = append(ctrs, c)
	}

	// Execute the template, passing any necessary data as a parameter
	err = tmpl.Execute(w, ctrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	// Define the route to handle requests for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getIndexHandler(w, r)
		}
	})

	// Start the server and listen for incoming requests
	fmt.Println("Listening on http://localhost:5555 ...")
	go http.ListenAndServe(":5555", nil)

	select {}
}
