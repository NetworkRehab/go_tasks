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
	"image/color"
	"log"
	"strconv"
	"time"
)

// Constants for UI configuration
const (
	WindowWidth  = 900
	WindowHeight = 1024
	MinWidth     = 430
	MinHeight    = 480
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

type customTheme struct{}

func (m *customTheme) BackgroundColor() color.Color {
	return color.Black
}

func (m *customTheme) ButtonColor() color.Color {
	return color.Gray{Y: 230}
}

func (m *customTheme) DisabledButtonColor() color.Color {
	return color.Gray{Y: 179}
}

func (m *customTheme) HyperlinkColor() color.Color {
	return color.RGBA{R: 0, G: 0, B: 255, A: 255}
}

func (m *customTheme) TextColor() color.Color {
	return color.RGBA{R: 0, G: 150, B: 255, A: 255}
}

func (m *customTheme) DisabledTextColor() color.Color {
	return color.Gray{Y: 128}
}

func (m *customTheme) IconColor() color.Color {
	return color.Black
}

func (m *customTheme) DisabledIconColor() color.Color {
	return color.Gray{Y: 128}
}

func (m *customTheme) PlaceHolderColor() color.Color {
	return color.Gray{Y: 128}
}

func (m *customTheme) PrimaryColor() color.Color {
	return color.RGBA{R: 0, G: 122, B: 255, A: 255}
}

func (m *customTheme) HoverColor() color.Color {
	return color.Gray{Y: 204} // 0.8 * 255 = 204
}

func (m *customTheme) FocusColor() color.Color {
	return color.RGBA{R: 0, G: 122, B: 255, A: 255}
}
func (m *customTheme) ScrollBarColor() color.Color {
	return color.Gray{Y: uint8(0.6 * 255)}
}

func (m *customTheme) ShadowColor() color.Color {
	return color.Gray{Y: uint8(0.4 * 255)}
}

func (m *customTheme) TextSize() int {
	return 14
}

func (m *customTheme) TextFont() fyne.Resource {
	return theme.DefaultTextFont()
}

func (m *customTheme) TextBoldFont() fyne.Resource {
	return theme.DefaultTextBoldFont()
}

func (m *customTheme) TextItalicFont() fyne.Resource {
	return theme.DefaultTextItalicFont()
}

func (m *customTheme) TextBoldItalicFont() fyne.Resource {
	return theme.DefaultTextBoldItalicFont()
}

func (m *customTheme) TextMonospaceFont() fyne.Resource {
	return theme.DefaultTextMonospaceFont()
}

func (m *customTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTextMonospaceFont()
	}
	if s.Bold {
		if s.Italic {
			return theme.DefaultTextBoldItalicFont()
		}
		return theme.DefaultTextBoldFont()
	}
	if s.Italic {
		return theme.DefaultTextItalicFont()
	}
	return theme.DefaultTextFont()
}

func (m *customTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (m *customTheme) Padding() int {
	return 8
}

func (m *customTheme) IconInlineSize() int {
	return 20
}

func (m *customTheme) ScrollBarSize() int {
	return 16
}

func (m *customTheme) ScrollBarSmallSize() int {
	return 3
}

func (m *customTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}

func (m *customTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return m.BackgroundColor()
	case theme.ColorNameButton:
		return m.ButtonColor()
	case theme.ColorNameDisabled:
		return m.DisabledButtonColor()
	case theme.ColorNameForeground:
		return m.TextColor()
	case theme.ColorNamePrimary:
		return m.PrimaryColor()
	case theme.ColorNameHover:
		return m.HoverColor()
	case theme.ColorNameScrollBar:
		return m.ScrollBarColor()
	case theme.ColorNameShadow:
		return m.ShadowColor()
	default:
		return theme.DefaultTheme().Color(n, v)
	}
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

	ctx = context.Background()
	if err := db.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

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
	App.Settings().SetTheme(&customTheme{}) // Set custom theme
	Window := App.NewWindow("Task Manager")

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
	pointsEntry.SetPlaceHolder("Enter points")

	// Create containers for tasks and completions
	tasksContainer := container.NewVBox()
	completionsContainer := container.NewVBox()

	tasksScroll := container.NewScroll(tasksContainer)
	tasksScroll.SetMinSize(fyne.NewSize(400, 760))

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
	completionsScroll.SetMinSize(fyne.NewSize(300, 900))

	// Add clear completions button
	var clearButton *widget.Button

	// Add task button with tooltips
	addButton := widget.NewButtonWithIcon("Add Task", theme.ContentAddIcon(), func() {
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
	split.SetOffset(0.5) // Left panel gets 60% of width

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
