package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAfterUserCreation(t *testing.T) {
	setup(t, "TestRunAfterUserCreation")
	defer teardown(t)
	config.RunAfterUserCreation = filepath.Join(testdataDir, "run-after-user.bat")
	PrepareTaskEnvironment()
	defer taskCleanup()
	file := filepath.Join(taskContext.TaskDir, "run-after-user.txt")
	_, err := os.Stat(file)
	if err != nil {
		t.Fatalf("Got error when looking for file %v: %v", file, err)
	}
}
