[![Build and Release](./assets/buildandrelease.jpg)](https://github.com/ndonathan/go_tasks/actions/workflows/release.yml)

# Task Tracker 📝

## Introduction 🚀

Go Task Tracker is a powerful cross-platform desktop application built with Go and [Fyne](https://fyne.io/). This robust task management tool helps users efficiently organize, track, and complete tasks while maintaining a comprehensive history of achievements. Supporting macOS*, Linux, and Windows platforms, it delivers a consistent and intuitive experience across all operating systems.

*macOS support is currently in development and will be available at some point in the future due to complexities with Xcode. 


## Key Features 🎯

- ✨ **Task Management**
    - Create tasks with custom names and point values
    - Mark tasks as completed with timestamp tracking
    - Delete individual tasks as needed

- 📊 **History Tracking**
    - View detailed completion history
    - Track points earned over time
    - Remove specific completion records
    - Bulk clear all history with one click

- 🌍 **Platform Support**
    - Native support for macOS, Linux, and Windows
    - Consistent UI/UX across platforms
    - Automated cross-platform builds

## Getting Started 🛠️

### System Requirements

- [Go](https://golang.org/dl/) (version 1.20+)
- [Git](https://git-scm.com/downloads)
- Operating System: macOS, Linux, or Windows

### Quick Start Guide

1. **Clone the Repository**
     ```bash
     git clone https://github.com/ndonathan/go_tasks.git
     cd go_tasks
     ```

2. **Build for Your Platform**
     ```bash
     # Default build
     go build

     # Platform-specific builds
     macOS:   GOOS=darwin GOARCH=amd64 go build
     Linux:   GOOS=linux GOARCH=amd64 go build
     Windows: GOOS=windows GOARCH=amd64 go build
     ```

3. **Launch the Application**
     ```bash
     # macOS/Linux
     ./go_tasks

     # Windows
     go_tasks.exe
     ```

## Usage Guide 📖

### Task Management
1. **Adding Tasks**
     - Enter task name
     - Set point value
     - Click "Add Task"

2. **Completing Tasks**
     - Select task
     - Click "Complete"
     - Confirm action

3. **Managing History**
     - View completions in history section
     - Delete individual records
     - Clear entire history

## Automated CI/CD Pipeline 🔄

### Release Types
- **Beta Releases (v0.x.x-beta)**
    - Triggered by pull requests
    - Perfect for testing new features

- **Production Releases (v1.x.x)**
    - Triggered by main branch pushes
    - Stable, tested versions

### Build Process
1. Automated compilation for all platforms
2. Asset packaging and bundling
3. Automatic release creation
4. Binary distribution

## Contributing 🤝

We welcome contributions! Follow these steps:

1. Fork repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

### Development Guidelines
- Follow Go best practices
- Include unit tests
- Update documentation
- Maintain cross-platform compatibility

## License & Credits 📝

- Licensed under MIT License
- Built with [Fyne](https://fyne.io/) UI toolkit
- Special thanks to all contributors