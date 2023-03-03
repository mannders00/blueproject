package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Data struct {
	Loc   string
	Price string
}

var d Data

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the index.html template file
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template, passing any necessary data as a parameter
	err = tmpl.Execute(w, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postIndexHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	location := r.FormValue("locationInput")
	price := r.FormValue("priceInput")
	d = Data{
		Loc:   location,
		Price: price,
	}
	getIndexHandler(w, r)
}

func main() {

	// Define the route to handle requests for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getIndexHandler(w, r)
		case "POST":
			postIndexHandler(w, r)
		}
	})

	// Start the server and listen for incoming requests
	fmt.Println("Listening on http://localhost:5555 ...")
	go http.ListenAndServe(":5555", nil)

	select {}
}
