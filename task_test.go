// Contains unit tests for the Task struct and its methods.

package main

import (
	"testing"
)

// TestAddTask tests the AddTask function.
func TestAddTask(t *testing.T) {
	db := NewDatabase()
	defer db.Conn.Close()
	db.Initialize()

	task, err := AddTask(db, "Test Task", 10)
	if err != nil {
		t.Fatal("Failed to add task:", err)
	}

	if task.Name != "Test Task" {
		t.Errorf("Expected task name 'Test Task', got '%s'", task.Name)
	}

	if task.Score != 10 {
		t.Errorf("Expected task score 10, got %d", task.Score)
	}
}

// TestTaskComplete tests the Complete method of Task.
func TestTaskComplete(t *testing.T) {
	db := NewDatabase()
	defer db.Conn.Close()
	db.Initialize()

	task, err := AddTask(db, "Test Task", 10)
	if err != nil {
		t.Fatal("Failed to add task:", err)
	}

	err = task.Complete(db)
	if err != nil {
		t.Fatal("Failed to complete task:", err)
	}

	// Additional assertions can be added to verify the completion was logged.
}

// TestEditScore tests the EditScore method of Task.
func TestEditScore(t *testing.T) {
	db := NewDatabase()
	defer db.Conn.Close()
	db.Initialize()

	task, err := AddTask(db, "Test Task", 10)
	if err != nil {
		t.Fatal("Failed to add task:", err)
	}

	err = task.EditScore(db, 20)
	if err != nil {
		t.Fatal("Failed to edit task score:", err)
	}

	if task.Score != 20 {
		t.Errorf("Expected task score 20 after edit, got %d", task.Score)
	}
}
