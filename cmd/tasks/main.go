package main

import (
	"context"
	"net/http"
	"log"
	"time"
	"html/template"
	"strconv"
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

func main() {
	database, err := NewDatabase()
	if (err != nil) {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	if err := database.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	appState := &AppState{
		db:          database,
		tasks:       make([]*Task, 0),
		completions: make([]*Completion, 0),
	}

	// Initialize tasks and completions
	refreshData(appState)

	// Start the HTTP server
	runServer(appState)
}

func runServer(appState *AppState) {
	// Serve static HTML
	http.HandleFunc("/", handleHome(appState))
	
	// Task endpoints
	http.HandleFunc("/tasks", handleTasks(appState))
	http.HandleFunc("/task/add", handleAddTask(appState))
	http.HandleFunc("/task/complete/", handleCompleteTask(appState))
	http.HandleFunc("/task/delete/", handleDeleteTask(appState))
	
	// Completion endpoints
	http.HandleFunc("/completions", handleCompletions(appState))
	http.HandleFunc("/completion/delete/", handleDeleteCompletion(appState))

	log.Printf("Server starting at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(_ *AppState) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>Tasks</title>
	<script src="https://unpkg.com/htmx.org@1.9.10"></script>
	<style>
		.completed { text-decoration: line-through; }
	</style>
</head>
<body>
	<h1>Tasks</h1>
	
	<div hx-get="/tasks" hx-trigger="load, taskChange from:body">
		<!-- Tasks load here -->
	</div>

	<form hx-post="/task/add" hx-trigger="submit" hx-target="#tasks">
		<input type="text" name="name" required>
		<input type="number" name="points" value="1">
		<button type="submit">Add Task</button>
	</form>

	<h2>Completions</h2>
	<div hx-get="/completions" hx-trigger="load, taskChange from:body">
		<!-- Completions load here -->
	</div>
</body>
</html>`
		writer.Write([]byte(tmpl))
	}
}

func handleTasks(appState *AppState) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        tasks, err := GetTasks(appState.db)
        if err != nil {
            http.Error(writer, err.Error(), 500)
            return
        }

        tmpl := `
        <div id="tasks">
            {{range .}}
            <div class="task">
                {{.Name}} ({{.Points}} pts)
                <button hx-post="/task/complete/{{.ID}}" 
                        hx-swap="none"
                        hx-trigger="click">Complete</button>
                <button hx-delete="/task/delete/{{.ID}}"
                        hx-swap="none"
                        hx-trigger="click">Delete</button>
            </div>
            {{end}}
        </div>`

        template := template.Must(template.New("tasks").Parse(tmpl))
        template.Execute(writer, tasks)
    }
}

func handleAddTask(appState *AppState) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := request.FormValue("name")
		points, _ := strconv.Atoi(request.FormValue("points"))
		
		_, err := appState.db.InsertTask(request.Context(), name, &points, "")
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}

		// Trigger refresh
		writer.Header().Set("HX-Trigger", "taskChange")
		writer.Write([]byte(""))
	}
}

func handleCompleteTask(appState *AppState) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if request.Method != "POST" {
            http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        taskID, err := strconv.Atoi(request.URL.Path[len("/task/complete/"):])
        if err != nil {
            http.Error(writer, "Invalid task ID", http.StatusBadRequest)
            return
        }

        if err := CompleteTask(request.Context(), appState.db, taskID); err != nil {
            http.Error(writer, err.Error(), http.StatusInternalServerError)
            return
        }

        writer.Header().Set("HX-Trigger", "taskChange")
        writer.Write([]byte(""))
    }
}

func handleDeleteTask(appState *AppState) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "DELETE" {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract task ID from URL path
		taskID, err := strconv.Atoi(request.URL.Path[len("/task/delete/"):])
		if err != nil {
			http.Error(writer, "Invalid task ID", http.StatusBadRequest)
			return
		}

		// Delete the task
		err = appState.db.DeleteTask(request.Context(), taskID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Trigger refresh
		writer.Header().Set("HX-Trigger", "taskChange")
		writer.Write([]byte(""))
	}
}

func handleCompletions(appState *AppState) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        completions, err := GetCompletions(appState.db)
        if err != nil {
            http.Error(writer, err.Error(), 500)
            return
        }

        tmpl := `
        <div id="completions">
            {{range .}}
            <div class="completion">
                {{.TaskName}} ({{.Points}} pts) - {{.CompletedAt.Format "2006-01-02 15:04"}}
                <button hx-delete="/completion/delete/{{.ID}}"
                        hx-swap="none"
                        hx-trigger="click">Delete</button>
            </div>
            {{end}}
        </div>`

        template := template.Must(template.New("completions").Parse(tmpl))
        template.Execute(writer, completions)
    }
}

func handleDeleteCompletion(appState *AppState) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "DELETE" {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract completion ID from URL path
		completionID, err := strconv.Atoi(request.URL.Path[len("/completion/delete/"):])
		if err != nil {
			http.Error(writer, "Invalid completion ID", http.StatusBadRequest)
			return
		}

		// Delete the completion
		err = appState.db.DeleteCompletion(request.Context(), completionID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Trigger refresh
		writer.Header().Set("HX-Trigger", "taskChange")
		writer.Write([]byte(""))
	}
}

// Additional handlers follow similar pattern

func refreshData(appState *AppState) {
	// Refresh tasks and completions
	tasks, err := GetTasks(appState.db)
	if err != nil {
		log.Printf("Failed to load tasks: %v", err)
		return
	}
	appState.tasks = tasks

	completions, err := GetCompletions(appState.db)
	if err != nil {
		log.Printf("Failed to load completions: %v", err)
		return
	}
	appState.completions = completions
}
