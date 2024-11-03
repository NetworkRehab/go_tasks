package main

import (
    "context"
    "database/sql"
    "time"
    _ "github.com/mattn/go-sqlite3"
)

type Database struct {
    Conn *sql.DB
}

func NewDatabase() (*Database, error) {
    // Remove os.Remove to prevent data loss on restart
    db, err := sql.Open("sqlite3", "./task_tracker.db?_timeout=5000&_journal_mode=WAL")
    if err != nil {
        return nil, err
    }

    // Test the connection
    if err := db.Ping(); err != nil {
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

    // Create tasks table
    if _, err := tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            points INTEGER NOT NULL,
            created_at DATETIME NOT NULL
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
