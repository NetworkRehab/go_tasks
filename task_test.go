// Contains unit tests for the Task struct and its methods.

package main

import (
	"context"
	"testing"
)

// TestAddTask tests the AddTask function.
func TestAddTask(t *testing.T) {
	db, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test invalid task
	_, err = AddTask(ctx, db, "", -1)
	if err == nil {
		t.Error("Expected error for invalid task, got nil")
	}

	// Test valid task
	task, err := AddTask(ctx, db, "Test Task", 10)
	if err != nil {
		t.Fatal("Failed to add task:", err)
	}

	if task.Name != "Test Task" {
		t.Errorf("Expected task name 'Test Task', got '%s'", task.Name)
	}

	if task.Points != 10 {
		t.Errorf("Expected task points 10, got %d", task.Points)
	}
}

// TestTaskCompletion ensures that completing a task adds a record
func TestTaskCompletion(t *testing.T) {
	db, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Conn.Close()
	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add a sample task
	task, err := AddTask(ctx, db, "Completion Test", 5)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Complete the task
	ctx = context.Background()
	err = CompleteTask(ctx, db, task.ID)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Verify completion exists
	var count int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM completions WHERE task_id = ?", task.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query completions: %v", err)
	}

	if count != 1 {
		t.Fatalf("Expected 1 completion, got %d", count)
	}
}

// TestClearCompletions verifies that ClearCompletions successfully deletes all completions.
func TestClearCompletions(t *testing.T) {
	db, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Conn.Close()
	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add and complete a task
	ctx = context.Background()
	task, err := AddTask(ctx, db, "Test Task", 10)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	err = CompleteTask(ctx, db, task.ID)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Verify completion exists
	var count int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM completions WHERE task_id = ?", task.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query completions: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 completion, got %d", count)
	}

	// Clear completions
	err = ClearCompletions(ctx, db)
	if err != nil {
		t.Fatalf("ClearCompletions failed: %v", err)
	}

	// Verify completions are cleared
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM completions WHERE task_id = ?", task.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query completions after clearing: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 completions after clearing, got %d", count)
	}
}

// TestClearCompletionsTableNotExist verifies behavior when the completions table does not exist.
func TestClearCompletionsTableNotExist(t *testing.T) {
	db, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Conn.Close()

	// Drop the completions table if exists
	_, err = db.Conn.Exec("DROP TABLE IF EXISTS completions")
	if err != nil {
		t.Fatalf("Failed to drop completions table: %v", err)
	}

	// Attempt to clear completions
	err = ClearCompletions(context.Background(), db)
	if err == nil {
		t.Fatalf("Expected error when clearing completions on non-existent table, got nil")
	}

	// Check error message
	expectedErrMsg := "task_completions table does not exist"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Fatalf("Expected error message to start with '%s', got '%s'", expectedErrMsg, err.Error())
	}
}