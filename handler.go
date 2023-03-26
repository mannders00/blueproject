package main

import (
	"encoding/json"
	"html/template"
	"net/http"
)

func (app *App) profileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		tmpl, err := template.New("index.html").ParseFiles("public/html/index.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		session, err := json.Marshal(getSession(request.Context()))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(writer, "index.html", string(session))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
