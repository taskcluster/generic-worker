package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAfterUserCreation(t *testing.T) {
	setup(t)
	if config.RunTasksAsCurrentUser {
		t.Skip("Skipping since running as current user...")
	}
	config.RunAfterUserCreation = []string{
		"C:\\Windows\\System32\\cmd.exe",
		"/c",
		"echo hello > %USERPROFILE%\run-after-user",
	}
	PrepareTaskEnvironment()
	file := filepath.Join(taskContext.TaskDir, "run-after-user")
	_, err := os.Stat(file)
	if err != nil {
		t.Fatal("Got error when looking for file %v: %v", file, err)
	}
}
