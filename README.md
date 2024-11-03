# Go Task Tracker

[![Build and Release](https://github.com/ndonathan/go_tasks/actions/workflows/release.yml/badge.svg)](https://github.com/ndonathan/go_tasks/actions/workflows/release.yml)

## Introduction

Go Task Tracker is a cross-platform desktop application built with Go and [Fyne](https://fyne.io/). It allows users to manage tasks efficiently by adding, completing, and deleting tasks, as well as tracking task completion history. The application supports macOS, Linux, and Windows, ensuring a seamless experience across different operating systems.

## Features

- **Add Tasks:** Create new tasks with a name and point value.
- **Complete Tasks:** Mark tasks as completed, automatically recording the completion time and points.
- **Track History:** View a history of all completed tasks.
- **Delete Tasks and Completions:** Remove tasks or specific completion records.
- **Clear History:** Remove all completion records with a single action.
- **Cross-Platform:** Available on macOS, Linux, and Windows.
- **Automated Builds and Releases:** Utilize GitHub Actions to build and release binaries for multiple platforms automatically.

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.20 or later installed on your machine.
- [Git](https://git-scm.com/downloads) installed.

### Clone the Repository
```bash
git clone https://github.com/ndonathan/go_tasks.git
cd go_tasks
```

Build the Application
You can build the application for your current operating system using the following command:

Building for Different Operating Systems
The project utilizes GitHub Actions to automatically build binaries for macOS, Linux, and Windows. To manually build for a specific OS, set the GOOS and GOARCH environment variables:

macOS:

Linux:

Windows:

Run the Application
After building, run the executable:

On macOS and Linux:

On Windows:

Usage
Add a New Task:

Enter the task name and points.
Click the "Add Task" button.
Complete a Task:

Click the "Complete" button next to the desired task.
Confirm the action in the dialog.
View Completion History:

Completed tasks appear in the "Completion History" section.
Delete Tasks or Completions:

Click the delete icon next to a task or completion to remove it.
Clear All Completions:

Click the "Clear History" button to remove all completion records.
Automated Builds and Releases
The project uses GitHub Actions to automate the build and release process for macOS, Linux, and Windows.

Workflow Details
Beta Releases:

Triggered when a pull request is opened against the main branch.
Releases are tagged with a major version of 0 (e.g., v0.1.0-beta).
General Releases:

Triggered when changes are pushed to the main branch.
Releases are tagged with a major version of 1 (e.g., v1.0.0).
Accessing Releases
You can find the released binaries under the Releases section of the repository.

Contributing
Contributions are welcome! Please follow these steps to contribute to the project:

Fork the Repository

Create a Feature Branch

Commit Your Changes

Push to the Branch

Open a Pull Request

Ensure that your code adheres to the project's coding standards and includes appropriate tests.

License
This project is licensed under the MIT License. See the LICENSE file for details.

Acknowledgments
Built with Fyne
Inspired by various task management tools.
Thanks to the contributors of GitHub Actions for enabling automated workflows.

