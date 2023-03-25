package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	var err error

	db, err := sql.Open("sqlite3", "db.db")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			verification_token TEXT NOT NULL UNIQUE,
			credits INTEGER NOT NULL DEFAULT 0,
			verified INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		panic(err)
	}
	return db
}
