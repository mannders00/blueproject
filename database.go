package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

func generateRandomString(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, (length+3)/4*3) // Ensure enough bytes are generated for Base64 encoding
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert bytes to a Base64 string and trim to desired length
	randomString := base64.RawURLEncoding.EncodeToString(bytes)
	return randomString[:length], nil
}

func InitDB() *sql.DB {

	db, err := sql.Open("sqlite3", "db.db")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY UNIQUE,
			user_id TEXT,
			data TEXT NOT NULL
		)
	`)
	if err != nil {
		panic(err)
	}

	return db

}

func (app *App) SaveProject(user_id string, project *ProjectPlan) (string, error) {

	data, err := json.Marshal(project)
	if err != nil {
		return "", err
	}

	unique_id, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	_, err = app.db.Exec("INSERT INTO projects (id, user_id, data) VALUES (?, ?, ?)", unique_id, user_id, string(data))
	if err != nil {
		return "", err
	}

	return unique_id, err

}

func (app *App) LoadProject(id string) (*ProjectPlan, error) {
	var data string
	err := app.db.QueryRow("SELECT data FROM projects WHERE id = ?", id).Scan(&data)
	if err != nil {
		return nil, err
	}

	objData := &ProjectPlan{}
	err = json.Unmarshal([]byte(data), &objData)
	if err != nil {
		return nil, err
	}

	return objData, nil
}
