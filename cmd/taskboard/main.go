// Command taskboard launches the terminal Kanban board from anywhere.
//
// Install once: go install ./cmd/taskboard   (ensures ~/go/bin is on PATH)
// Then run:    taskboard  |  taskboard -f FILE  |  taskboard run  |  taskboard version
package main

import (
	"fmt"
	"os"

	"kanban/internal/cli"
)

// version is the build version; override at link time, e.g.
// go build -ldflags "-X main.version=v1.0.0" -o bin/taskboard ./cmd/taskboard
var version = "dev"

func main() {
	if err := cli.Execute("taskboard", "taskboard — terminal Kanban board", version); err != nil {
		fmt.Fprintln(os.Stderr, "taskboard:", err)
		os.Exit(1)
	}
}
