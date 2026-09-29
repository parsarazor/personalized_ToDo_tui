// Command todo launches the terminal Kanban board from anywhere.
//
// Install once: go install ./cmd/todo   (ensures ~/go/bin is on PATH)
// Then run:    todo  |  todo -f FILE  |  todo run  |  todo version
package main

import (
	"fmt"
	"os"

	"kanban/internal/cli"
)

// version is the build version; override at link time, e.g.
// go build -ldflags "-X main.version=v1.0.0" -o bin/todo ./cmd/todo
var version = "dev"

func main() {
	if err := cli.Execute("todo", "todo — terminal Kanban board", version); err != nil {
		fmt.Fprintln(os.Stderr, "todo:", err)
		os.Exit(1)
	}
}
