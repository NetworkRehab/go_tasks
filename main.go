package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Initialize the Fyne application.
	myApp := app.New()
	myWindow := myApp.NewWindow("Task Manager")

	// Initialize the database connection.
	db := NewDatabase()
	defer db.Conn.Close()

	// Create the necessary database tables.
	db.Initialize()

	// Create input fields for adding a new task.
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter task name")

	scoreEntry := widget.NewEntry()
	scoreEntry.SetPlaceHolder("Enter task score")

	// Create a button to add tasks.
	addButton := widget.NewButton("Add Task", func() {
		name := nameEntry.Text
		score, err := strconv.Atoi(scoreEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid score: %v", err), myWindow)
			return
		}

		task, err := AddTask(db, name, score)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error adding task: %v", err), myWindow)
			return
		}

		// Update the task list.
		tasks, err := GetTasks(db)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error retrieving tasks: %v", err), myWindow)
			return
		}
		updateTaskList(taskList, tasks)

		// Show success message.
		dialog.ShowInformation("Success", fmt.Sprintf("Added Task: ID=%d, Name=%s, Score=%d", task.ID, task.Name, task.Score), myWindow)

		// Clear the input fields.
		nameEntry.SetText("")
		scoreEntry.SetText("")
	})

	// Create a list to display tasks.
	taskList := widget.NewList(
		func() int {
			tasks, _ := GetTasks(db)
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			tasks, _ := GetTasks(db)
			if i >= len(tasks) {
				o.(*widget.Label).SetText("")
				return
			}
			task := tasks[i]
			o.(*widget.Label).SetText(fmt.Sprintf("ID=%d | %s | Score: %d | Created: %s | Updated: %s",
				task.ID, task.Name, task.Score,
				task.CreatedAt.Format("2006-01-02"),
				task.UpdatedAt.Format("2006-01-02")))
		},
	)

	// Arrange widgets in the window.
	content := container.NewVBox(
		widget.NewLabel("Add New Task"),
		nameEntry,
		scoreEntry,
		addButton,
		widget.NewSeparator(),
		widget.NewLabel("Tasks"),
		taskList,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 600))
	myWindow.ShowAndRun()
}

// updateTaskList refreshes the task list widget with the latest tasks.
func updateTaskList(list *widget.List, tasks []*Task) {
	list.Length = func() int {
		return len(tasks)
	}
	list.Refresh()
}
