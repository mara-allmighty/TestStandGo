package main

import (
	"database/sql"
	"log"
	"os"
)

func PostgresConnection() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to open database: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	return db
}
