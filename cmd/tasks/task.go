// Defines the Task struct and associated methods.

package main

import (
	"context"
	"database/sql" // Add this import
	"fmt"
	"time"
)

type Task struct {
	ID        int
	Name      string
	Points    int
	Notes     string
	CreatedAt time.Time
}

func (t *Task) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if t.Points < 0 {
		return fmt.Errorf("points cannot be negative")
	}
	return nil
}

func AddTask(ctx context.Context, db *Database, name string, points *int, notes string) (*Task, error) {
	pointsValue := 0
	if points != nil {
		pointsValue = *points
	}

	task := &Task{
		Name:      name,
		Points:    pointsValue,
		Notes:     notes,
		CreatedAt: time.Now(),
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	tx, err := db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO tasks (name, points, notes, created_at) VALUES (?, ?, ?, ?)`,
		task.Name, task.Points, task.Notes, task.CreatedAt)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	task.ID = int(id)

	return task, tx.Commit()
}

func GetTasks(db *Database) ([]*Task, error) {
	// Only return non-deleted tasks
	query := `SELECT id, name, points, notes, created_at FROM tasks WHERE deleted = 0`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Name, &task.Points, &task.Notes, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func CompleteTask(ctx context.Context, db *Database, taskID int) error {
	tx, err := db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify task exists
	var points int
	err = tx.QueryRowContext(ctx, "SELECT points FROM tasks WHERE id = ?", taskID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("task not found: %d", taskID)
		}
		return err
	}

	// Record completion
	_, err = tx.ExecContext(ctx,
		`INSERT INTO completions (task_id, completed_at, points) VALUES (?, ?, ?)`,
		taskID, time.Now(), points)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetCompletions(db *Database) ([]*Completion, error) {
	query := `
        SELECT c.id, c.task_id, c.completed_at, c.points, 
               CASE WHEN t.deleted = 1 THEN t.name || ' (deleted)' ELSE t.name END as task_name
        FROM completions c
        LEFT JOIN tasks t ON c.task_id = t.id
        ORDER BY c.completed_at DESC
    `
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var completions []*Completion
	for rows.Next() {
		c := &Completion{}
		err := rows.Scan(&c.ID, &c.TaskID, &c.CompletedAt, &c.Points, &c.TaskName)
		if err != nil {
			return nil, err
		}
		completions = append(completions, c)
	}
	return completions, nil
}

// ClearCompletions removes all task completion records from the database and updates tasks if needed
func ClearCompletions(ctx context.Context, db *Database) error {
	tx, err := db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear the completions table
	_, err = tx.ExecContext(ctx, `DELETE FROM completions`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func DeleteTask(ctx context.Context, db *Database, taskID int) error {
	tx, err := db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Mark task as deleted instead of removing it
	result, err := tx.ExecContext(ctx, "UPDATE tasks SET deleted = 1 WHERE id = ?", taskID)
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

	return tx.Commit()
}

func DeleteCompletion(ctx context.Context, db *Database, completionID int) error {
	_, err := db.Conn.ExecContext(ctx, "DELETE FROM completions WHERE id = ?", completionID)
	return err
}

// Update function signature to match usage
func CreateTask(db *Database, name string, points *int, notes string) error {
    id, err := db.InsertTask(context.Background(), name, points, notes)
    if err != nil {
        return err
    }
    fmt.Printf("Task created with ID: %d\n", id)
    return nil
}

func UpdateTaskNotes(ctx context.Context, db *Database, taskID int, notes string) error {
    tx, err := db.Conn.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    result, err := tx.ExecContext(ctx, 
        "UPDATE tasks SET notes = ? WHERE id = ? AND deleted = 0", 
        notes, taskID)
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

    return tx.Commit()
}

// Add new function to update task points
func UpdateTaskPoints(ctx context.Context, db *Database, taskID int, points int) error {
    if points < 0 {
        return fmt.Errorf("points cannot be negative")
    }

    tx, err := db.Conn.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    result, err := tx.ExecContext(ctx, 
        "UPDATE tasks SET points = ? WHERE id = ? AND deleted = 0", 
        points, taskID)
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

    return tx.Commit()
}
