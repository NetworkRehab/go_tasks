// Defines the Task struct and associated methods.

package main

import (
	"time"
)

// Task represents a task that can be completed multiple times.
type Task struct {
	ID        int       // Unique identifier for the task.
	Name      string    // Name of the task.
	Score     int       // Editable score of the task.
	CreatedAt time.Time // Timestamp when the task was created.
	UpdatedAt time.Time // Timestamp when the task was last updated.
}

// Complete logs the completion of the task.
func (t *Task) Complete(db *Database) error {
	// Insert a record into the task_completions table with the current timestamp and score.
	query := `
        INSERT INTO task_completions (task_id, completed_at, score)
        VALUES (?, ?, ?)
    `
	_, err := db.Conn.Exec(query, t.ID, time.Now(), t.Score)
	return err
}

// EditScore updates the score of the task.
func (t *Task) EditScore(db *Database, newScore int) error {
	t.Score = newScore
	t.UpdatedAt = time.Now()

	// Update the task's score in the database.
	query := `
        UPDATE tasks SET score = ?, updated_at = ?
        WHERE id = ?
    `
	_, err := db.Conn.Exec(query, t.Score, t.UpdatedAt, t.ID)
	return err
}

// AddTask adds a new task to the database.
func AddTask(db *Database, name string, score int) (*Task, error) {
	createdAt := time.Now()
	updatedAt := createdAt

	// Insert the new task into the tasks table.
	query := `
        INSERT INTO tasks (name, score, created_at, updated_at)
        VALUES (?, ?, ?, ?)
    `
	result, err := db.Conn.Exec(query, name, score, createdAt, updatedAt)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	task := &Task{
		ID:        int(id),
		Name:      name,
		Score:     score,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return task, nil
}

// GetTasks retrieves all tasks from the database.
func GetTasks(db *Database) ([]*Task, error) {
	// Query the tasks table to retrieve all tasks.
	query := `
        SELECT id, name, score, created_at, updated_at
        FROM tasks
    `
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Name, &task.Score, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
