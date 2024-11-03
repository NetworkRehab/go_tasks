package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver.
)

// Database represents the SQLite database connection.
type Database struct {
	Conn *sql.DB // Database connection object.
}

// NewDatabase initializes a new database connection.
func NewDatabase() *Database {
	// Open a connection to the SQLite database file.
	db, err := sql.Open("sqlite3", "./task_tracker.db")
	if err != nil {
		log.Fatal(err)
	}
	return &Database{Conn: db}
}

// Initialize creates the necessary tables if they don't exist.
func (db *Database) Initialize() {
	// Create the tasks table to store task details.
	createTasksTable := `
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            score INTEGER NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL
        );
    `
	_, err := db.Conn.Exec(createTasksTable)
	if err != nil {
		log.Fatal(err)
	}

	// Create the task_completions table to log task completions.
	createCompletionsTable := `
        CREATE TABLE IF NOT EXISTS task_completions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER NOT NULL,
            completed_at DATETIME NOT NULL,
            score INTEGER NOT NULL,
            FOREIGN KEY(task_id) REFERENCES tasks(id)
        );
    `
	_, err = db.Conn.Exec(createCompletionsTable)
	if err != nil {
		log.Fatal(err)
	}
}
