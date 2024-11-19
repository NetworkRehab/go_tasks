package main

import (
    "context"
    "database/sql"
    "time"
    "os"
    "fmt"  // Add missing import
    _ "github.com/mattn/go-sqlite3"
)

type Database struct {
    Conn *sql.DB
}

func NewDatabase() (*Database, error) {
    // Ensure sqlite_db directory exists
    if err := os.MkdirAll("./sqlite_db", 0755); err != nil {
        return nil, fmt.Errorf("failed to create database directory: %v", err)
    }

    // Updated database path to store in sqlite_db directory
    db, err := sql.Open("sqlite3", "./sqlite_db/task_tracker.db?_timeout=5000&_journal_mode=WAL")
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Test the connection
    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return nil, err
    }

    // Set connection pool settings
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(time.Hour)

    return &Database{Conn: db}, nil
}

func (db *Database) Close() error {
    return db.Conn.Close()
}

func (db *Database) Initialize(ctx context.Context) error {
    // Create tables within transaction
    tx, err := db.Conn.BeginTx(ctx, nil)
    if (err != nil) {
        return err
    }
    defer tx.Rollback()

    // Create tasks table with deleted flag and notes
    if _, err := tx.ExecContext(ctx, `
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
    if _, err := tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS completions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER NOT NULL,
            completed_at DATETIME NOT NULL,
            points INTEGER NOT NULL,
            FOREIGN KEY(task_id) REFERENCES tasks(id)
        );`); err != nil {
        return err
    }

    return tx.Commit()
}

// Add migration function to add deleted column if it doesn't exist
func (db *Database) Migrate(ctx context.Context) error {
    // Check if deleted column exists
    var count int
    err := db.Conn.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='deleted'
    `).Scan(&count)
    if err != nil {
        return err
    }

    // Add column if it doesn't exist
    if count == 0 {
        _, err = db.Conn.ExecContext(ctx, `
            ALTER TABLE tasks ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT 0
        `)
        if err != nil {
            return err
        }
    }

    // Add notes column if it doesn't exist
    err = db.Conn.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='notes'
    `).Scan(&count)
    if err != nil {
        return err
    }

    if count == 0 {
        _, err = db.Conn.ExecContext(ctx, `
            ALTER TABLE tasks ADD COLUMN notes TEXT
        `)
        if err != nil {
            return err
        }
    }

    return nil
}

// Modify InsertTask to include notes
func (db *Database) InsertTask(ctx context.Context, name string, points *int, notes string) (int64, error) {
    var pts int
    if points != nil {
        pts = *points
    } else {
        pts = 0 // Default to 0 if not provided
    }

    result, err := db.Conn.ExecContext(ctx, `
        INSERT INTO tasks (name, points, notes, created_at)
        VALUES (?, ?, ?, ?)`,
        name, pts, notes, time.Now(),
    )
    if err != nil {
        return 0, err
    }

    return result.LastInsertId()
}
