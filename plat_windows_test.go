package main

import "testing"

// Test APPDATA / LOCALAPPDATA folder are not shared between tasks
func TestAppDataNotShared(t *testing.T) {
	// Assuming they are not reused, both tasks should be able to write to
	// them. If they are reused, the second task should fail since it does not
	// have permission to write to the file that is owned by the previous user.
	setup(t)
	payload := GenericWorkerPayload{
		Command: []string{
			"echo hello > %APPDATA%\\hello.txt",
			"echo hello > %LOCALAPPDATA%\\hello.txt",
		},
		MaxRunTime: 10,
	}
	td := testTask()

	// submit twice, so we see if there is a problem for the second task
	taskID1, myQueue := submitTask(t, td, payload)
	taskID2, _ := submitTask(t, td, payload)

	config.NumberOfTasksToRun = 2
	runWorker()

	tsr, err := myQueue.Status(taskID2)
	if err != nil {
		t.Fatalf("Could not retrieve task status")
	}
	if tsr.Status.State != "succeeded" {
		t.Fatalf("Was expecting state %q but got %q", "succeeded", tsr.Status.State)
	}
}
