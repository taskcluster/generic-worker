// +build docker

package dockerworker

import (
	"fmt"
	"io"
	"log"
	"os"
)

// NewLogger returns a new Logger that will write to standard output
func NewLogger(taskID string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[generic-docker-worker taskID=\"%s\"] ", taskID), 0)
}

// NewTaskLogger returns a logger that will writer to the task log
func NewTaskLogger(writer io.Writer) *log.Logger {
	return log.New(writer, "", 0)
}
