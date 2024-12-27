package main

import (
    "context"
    "database/sql"
    "time"
    "os"
    "fmt"
    _ "modernc.org/sqlite" // Changed import from mattn/go-sqlite3
)

type Database struct {
    Conn *sql.DB
}

func NewDatabase() (*Database, error) {
    // Ensure sqlite_db directory exists
    if err := os.MkdirAll("../sqlite_db", 0755); err != nil {
        return nil, fmt.Errorf("failed to create database directory: %v", err)
    }

    // Remove WAL mode as it's not needed with modernc/sqlite
    databaseConnection, err := sql.Open("sqlite", "../sqlite_db/task_tracker.db")
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Test the connection
    if err := databaseConnection.PingContext(ctx); err != nil {
        databaseConnection.Close()
        return nil, err
    }

    // Set connection pool settings
    databaseConnection.SetMaxOpenConns(10)
    databaseConnection.SetMaxIdleConns(5)
    databaseConnection.SetConnMaxLifetime(time.Hour)

    return &Database{Conn: databaseConnection}, nil
}

func (db *Database) Close() error {
    return db.Conn.Close()
}

func (db *Database) Initialize(ctx context.Context) error {
    // Create tables within transaction
    transaction, err := db.Conn.BeginTx(ctx, nil)
    if (err != nil) {
        return err
    }
    defer transaction.Rollback()

    // Create tasks table with deleted flag and notes
    if _, err := transaction.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            points INTEGER NOT NULL DEFAULT 0,
            notes TEXT,
            created_at DATETIME NOT NULL,
            deleted BOOLEAN NOT NULL DEFAULT 0
        );`); err != nil {
        return err
    }

    // Create completions table
    if _, err := transaction.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS completions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER NOT NULL,
            completed_at DATETIME NOT NULL,
            points INTEGER NOT NULL,
            FOREIGN KEY(task_id) REFERENCES tasks(id)
        );`); err != nil {
        return err
    }

    return transaction.Commit()
}

// Add migration function to add deleted column if it doesn't exist
func (db *Database) Migrate(ctx context.Context) error {
    // Ensure all necessary migrations are applied
    migrations := []struct {
        query string
    }{
        {
            query: `
                CREATE TABLE IF NOT EXISTS tasks (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    name TEXT NOT NULL,
                    points INTEGER NOT NULL DEFAULT 0,
                    notes TEXT,
                    created_at DATETIME NOT NULL,
                    deleted BOOLEAN NOT NULL DEFAULT 0
                );`,
        },
        {
            query: `
                CREATE TABLE IF NOT EXISTS completions (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    task_id INTEGER NOT NULL,
                    completed_at DATETIME NOT NULL,
                    points INTEGER NOT NULL,
                    FOREIGN KEY(task_id) REFERENCES tasks(id)
                );`,
        },
        // Add more migrations as needed
    }

    transaction, err := db.Conn.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer transaction.Rollback()

    for _, migration := range migrations {
        _, err := transaction.ExecContext(ctx, migration.query)
        if err != nil {
            return err
        }
    }

    return transaction.Commit()
}

// Modify InsertTask to include notes
func (db *Database) InsertTask(ctx context.Context, name string, points *int, notes string) (int64, error) {
    var taskPoints int
    if points != nil {
        taskPoints = *points
    } else {
        taskPoints = 0 // Default to 0 if not provided
    }

    result, err := db.Conn.ExecContext(ctx, `
        INSERT INTO tasks (name, points, notes, created_at)
        VALUES (?, ?, ?, ?)`,

        name, taskPoints, notes, time.Now(),
    )
    if err != nil {
        return 0, err
    }

    return result.LastInsertId()
}

// Add DeleteTask method to Database struct
func (db *Database) DeleteTask(ctx context.Context, taskID int) error {
    result, err := db.Conn.ExecContext(ctx, 
        "UPDATE tasks SET deleted = 1 WHERE id = ?", 
        taskID)
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return fmt.Errorf("task not found: %d", taskID)
    }
    return nil
}

// Add DeleteCompletion method to Database struct
func (db *Database) DeleteCompletion(ctx context.Context, completionID int) error {
    _, err := db.Conn.ExecContext(ctx, 
        "DELETE FROM completions WHERE id = ?", 
        completionID)
    return err
}
