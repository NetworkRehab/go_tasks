package main

import (
	"context" // Added missing import
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"strconv"
	"time"
)

// Constants for UI configuration
const (
	WindowWidth  = 1024
	WindowHeight = 768
	MinWidth     = 800
	MinHeight    = 600
)

type AppState struct {
	db          *Database
	tasks       []*Task
	completions []*Completion
}

type Completion struct {
	ID          int
	TaskID      int
	CompletedAt time.Time
	Points      int
	TaskName    string
}

func main() {
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Example: Insert a new task
	taskID, err := db.InsertTask(ctx, "Sample Task", nil)
	if err != nil {
		log.Fatalf("Failed to insert task: %v", err)
	}
	log.Printf("Inserted task with ID: %d", taskID)

	// ...additional code...

	// Initialize database with error handling
	db, err = NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	if err := db.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	state := &AppState{
		db:          db,
		tasks:       make([]*Task, 0),
		completions: make([]*Completion, 0),
	}

	// Initialize Fyne application
	App := app.New()
	Window := App.NewWindow("Task Tracker")

	// Set up error recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			dialog.ShowError(fmt.Errorf("unexpected error occurred"), Window)
		}
	}()

	// Create UI components
	ui := createUI(Window, state)

	// Set up window properties
	Window.SetContent(ui)
	Window.Resize(fyne.NewSize(WindowWidth, WindowHeight))
	Window.SetFixedSize(false)
	Window.CenterOnScreen()

	// Add keyboard shortcuts
	setupKeyboardShortcuts(Window, state)

	// Initial data load
	refreshData(ctx, state)

	Window.ShowAndRun()
}

// Fix createUI function to prevent window recreation
func createUI(window fyne.Window, state *AppState) fyne.CanvasObject {
	// Create input fields for adding a new task.
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter task name")

	pointsEntry := widget.NewEntry()
	pointsEntry.SetPlaceHolder("Enter points (0 or more)")

	// Create containers for tasks and completions
	tasksContainer := container.NewVBox()
	completionsContainer := container.NewVBox()

	tasksScroll := container.NewScroll(tasksContainer)
	tasksScroll.SetMinSize(fyne.NewSize(300, 500))

	// Helper function to show confirmation dialog
	showConfirmDialog := func(title, message string, callback func()) {
		dialog.ShowConfirm(title, message, func(ok bool) {
			if ok {
				callback()
			}
		}, window)
	}

	// Add task validation
	validateTask := func(name string, points int) error {
		if name == "" {
			return fmt.Errorf("task name cannot be empty")
		}
		if points < 0 {
			return fmt.Errorf("points cannot be negative")
		}
		return nil
	}

	// Declare update functions
	var updateTasks func()
	var updateCompletions func()

	updateCompletions = func() {
		completionsContainer.Objects = nil // Clear existing items
		completions, _ := GetCompletions(state.db)

		for _, c := range completions {
			// Create a styled completion entry
			dateStr := c.CompletedAt.Format("Jan 2, 2006")
			timeStr := c.CompletedAt.Format("3:04 PM")

			// Create delete button
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			deleteBtn.OnTapped = func(completionID int) func() {
				return func() {
					showConfirmDialog("Delete Completion",
						"Are you sure you want to delete this completion record?",
						func() {
							ctx := context.Background()
							if err := DeleteCompletion(ctx, state.db, completionID); err != nil {
								dialog.ShowError(err, window)
								return
							}
							updateCompletions()
						})
				}
			}(c.ID)

			// Update entry card to include delete button
			entryCard := widget.NewCard("", "", container.NewVBox(
				widget.NewLabelWithStyle(c.TaskName,
					fyne.TextAlignLeading,
					fyne.TextStyle{Bold: true}),
				container.NewHBox(
					widget.NewLabelWithStyle(fmt.Sprintf("Completed: %s at %s", dateStr, timeStr),
						fyne.TextAlignLeading,
						fyne.TextStyle{Italic: true}),
					layout.NewSpacer(),
					widget.NewLabelWithStyle(fmt.Sprintf("%d points", c.Points),
						fyne.TextAlignTrailing,
						fyne.TextStyle{Bold: true}),
					deleteBtn,
				),
			))

			completionsContainer.Add(entryCard)
		}
	}

	updateTasks = func() {
		tasksContainer.Objects = nil // Clear existing items
		tasks, err := GetTasks(state.db)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load tasks: %v", err), window)
			return
		}

		for _, task := range tasks {
			// Create a styled task entry with hover effect
			completeBtn := widget.NewButton("Complete", nil)
			completeBtn.Importance = widget.HighImportance

			completeBtn.OnTapped = func(taskID int, taskName string) func() {
				return func() {
					showConfirmDialog("Complete Task",
						fmt.Sprintf("Complete task '%s'?", taskName),
						func() {
							ctx := context.Background()
							if err := CompleteTask(ctx, state.db, taskID); err != nil {
								dialog.ShowError(err, window)
								return
							}
							updateTasks()
							updateCompletions()
						})
				}
			}(task.ID, task.Name)

			// Color based on points
			pointsLabel := widget.NewLabelWithStyle(
				fmt.Sprintf("%d points", task.Points),
				fyne.TextAlignLeading,
				fyne.TextStyle{Italic: true},
			)
			if task.Points >= 10 {
				pointsLabel.TextStyle.Bold = true
			}

			// Create delete button
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			deleteBtn.OnTapped = func(taskID int, taskName string) func() {
				return func() {
					showConfirmDialog("Delete Task",
						fmt.Sprintf("Are you sure you want to delete task '%s'?", taskName),
						func() {
							ctx := context.Background()
							if err := DeleteTask(ctx, state.db, taskID); err != nil {
								dialog.ShowError(err, window)
								return
							}
							updateTasks()
							updateCompletions()
						})
				}
			}(task.ID, task.Name)

			// Update task card to include delete button
			taskCard := widget.NewCard("", "", container.NewVBox(
				widget.NewLabelWithStyle(task.Name,
					fyne.TextAlignLeading,
					fyne.TextStyle{Bold: true}),
				container.NewHBox(
					pointsLabel,
					layout.NewSpacer(),
					completeBtn,
					deleteBtn,
				),
			))

			tasksContainer.Add(taskCard)
		}
	}

	// Create a scrollable container for completions
	completionsScroll := container.NewScroll(completionsContainer)
	completionsScroll.SetMinSize(fyne.NewSize(300, 400))

	// Add clear completions button
	clearButton := widget.NewButton("Clear History", func() {
		showConfirmDialog("Clear History",
			"Are you sure you want to clear all completion history?",
			func() {
				ctx := context.Background()
				if err := ClearCompletions(ctx, state.db); err != nil {
					dialog.ShowError(err, window)
					return
				}
				updateCompletions()
			})
	})

	// Add task button
	addButton := widget.NewButton("Add Task", func() {
		var points *int
		if pointsEntry.Text != "" {
			p, err := strconv.Atoi(pointsEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid points value"), window)
				return
			}
			points = &p
		}
		
		pointsValue := 0
		if points != nil {
			pointsValue = *points
		}
		
		if err := validateTask(nameEntry.Text, pointsValue); err != nil {
			dialog.ShowError(err, window)
			return
		}
		
		_, err := AddTask(context.Background(), state.db, nameEntry.Text, &pointsValue)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
	
		nameEntry.SetText("")
		pointsEntry.SetText("")
		updateTasks()
	})

	// Add tooltips
	addButton = widget.NewButtonWithIcon("Add Task", theme.ContentAddIcon(), func() {
		var points *int
		if pointsEntry.Text != "" {
			p, err := strconv.Atoi(pointsEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid points value"), window)
				return
			}
			points = &p
		}

		pointsValue := 0
		if points != nil {
			pointsValue = *points
		}
		if err := validateTask(nameEntry.Text, pointsValue); err != nil {
			dialog.ShowError(err, window)
			return
		}

		_, err := AddTask(context.Background(), state.db, nameEntry.Text, points)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		nameEntry.SetText("")
		pointsEntry.SetText("")
		updateTasks()
	})
	addButton.Importance = widget.HighImportance

	clearButton = widget.NewButtonWithIcon("Clear History", theme.ContentClearIcon(), func() {
		showConfirmDialog("Clear History",
			"Are you sure you want to clear all completion history?",
			func() {
				ctx := context.Background()
				if err := ClearCompletions(ctx, state.db); err != nil {
					dialog.ShowError(err, window)
					return
				}
				updateCompletions()
			})
	})
	clearButton.Importance = widget.HighImportance

	// Create scrolling containers with proper sizing
	leftContent := container.NewVBox(
		widget.NewCard("Add New Task", "",
			container.NewVBox(
				nameEntry,
				pointsEntry,
				addButton,
			),
		),
		widget.NewCard("Active Tasks", "",
			container.NewPadded(tasksScroll),
		),
	)

	// Modify right content to include clear button
	rightContent := container.NewVBox(
		widget.NewCard("Completion History", "",
			container.NewVBox(
				container.NewHBox(
					clearButton,
					layout.NewSpacer(),
				),
				container.NewPadded(completionsScroll),
			),
		),
	)

	// Initial load of tasks and completions
	updateTasks()
	updateCompletions()

	// Use HSplit with a specific divider position
	split := container.NewHSplit(
		container.NewPadded(leftContent),
		container.NewPadded(rightContent),
	)
	split.SetOffset(0.4) // Left panel gets 60% of width

	// Make window larger and set minimum size
	window.Resize(fyne.NewSize(1024, 768))
	window.SetContent(split)

	// Return the split container as the main UI
	return split
}

func setupKeyboardShortcuts(w fyne.Window, state *AppState) {
	shortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeyN,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(shortcut, func(shortcut fyne.Shortcut) {
		// Focus name entry
		// You can add the name entry focus logic here
	})
}

func refreshData(ctx context.Context, state *AppState) {
	// Refresh tasks and completions
	tasks, err := GetTasks(state.db)
	if err != nil {
		log.Printf("Failed to load tasks: %v", err)
		return
	}
	state.tasks = tasks

	completions, err := GetCompletions(state.db)
	if err != nil {
		log.Printf("Failed to load completions: %v", err)
		return
	}
	state.completions = completions
}
