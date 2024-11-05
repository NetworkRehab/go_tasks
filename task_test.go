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

	tests := []struct {
		name        string
		taskName    string
		points      *int
		shouldError bool
	}{
		{
			name:        "Valid task with points",
			taskName:    "Test Task",
			points:      intPtr(10),
			shouldError: false,
		},
		{
			name:        "Valid task without points",
			taskName:    "No Points Task",
			points:      nil,
			shouldError: false,
		},
		{
			name:        "Empty name",
			taskName:    "",
			points:      intPtr(5),
			shouldError: true,
		},
		{
			name:        "Negative points",
			taskName:    "Negative Points",
			points:      intPtr(-1),
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := AddTask(ctx, db, tt.taskName, tt.points)
			
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if task.Name != tt.taskName {
				t.Errorf("Expected task name %s, got %s", tt.taskName, task.Name)
			}

			expectedPoints := 0
			if tt.points != nil {
				expectedPoints = *tt.points
			}

			if task.Points != expectedPoints {
				t.Errorf("Expected task points %d, got %d", expectedPoints, task.Points)
			}
		})
	}
}

// Helper function to create integer pointer
func intPtr(i int) *int {
	return &i
}

// TestTaskCompletion ensures that completing a task adds a record
func TestTaskCompletion(t *testing.T) {
	db, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test with both nil and non-nil points
	tests := []struct {
		name   string
		points *int
	}{
		{
			name:   "Complete task with points",
			points: intPtr(5),
		},
		{
			name:   "Complete task without points",
			points: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add a sample task
			task, err := AddTask(ctx, db, "Completion Test", tt.points)
			if err != nil {
				t.Fatalf("Failed to add task: %v", err)
			}

			// Complete the task
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
				t.Errorf("Expected 1 completion, got %d", count)
			}

			// Verify points were recorded correctly
			var completionPoints int
			err = db.Conn.QueryRow("SELECT points FROM completions WHERE task_id = ?", task.ID).Scan(&completionPoints)
			if err != nil {
				t.Fatalf("Failed to query completion points: %v", err)
			}

			expectedPoints := 0
			if tt.points != nil {
				expectedPoints = *tt.points
			}

			if completionPoints != expectedPoints {
				t.Errorf("Expected completion points %d, got %d", expectedPoints, completionPoints)
			}
		})
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
	task, err := AddTask(ctx, db, "Test Task", intPtr(10))
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
