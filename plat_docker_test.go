// +build docker

package main

import (
	"testing"

	"github.com/taskcluster/generic-worker/dockerworker"
)

func TestSimpleTask(t *testing.T) {
	defer setup(t)()

	const message = "Lamb is watching"

	payload := dockerworker.DockerWorkerPayload{
		Command:    []string{"/bin/echo", message},
		MaxRunTime: 30,
	}

	td := testTask(t)

	expectedArtifacts := ExpectedArtifacts{
		"public/logs/live.log": {
			Extracts:        []string{message},
			ContentType:     "text/plain; charset=utf-8",
			ContentEncoding: "gzip",
			Expires:         td.Expires,
		},
	}

	taskID := scheduleDockerTask(t, td, payload)
	ensureResolution(t, taskID, "completed", "completed")

	expectedArtifacts.Validate(t, taskID, 0)
}

func TestEnvVar(t *testing.T) {
	defer setup(t)()

	const message = "I chose the impossible"
	const varName = "MY_MESSAGE"

	payload := dockerworker.DockerWorkerPayload{
		Env: map[string]string{
			varName: message,
		},
		Command:    []string{"/usr/bin/env"},
		MaxRunTime: 30,
	}

	td := testTask(t)

	expectedArtifacts := ExpectedArtifacts{
		"public/logs/live.log": {
			Extracts: []string{
				varName,
				message,
			},
			ContentType:     "text/plain; charset=utf-8",
			ContentEncoding: "gzip",
			Expires:         td.Expires,
		},
	}

	taskID := scheduleDockerTask(t, td, payload)
	ensureResolution(t, taskID, "completed", "completed")

	expectedArtifacts.Validate(t, taskID, 0)
}

func TestFailed(t *testing.T) {
	defer setup(t)()

	payload := dockerworker.DockerWorkerPayload{
		Command:    []string{"/bin/false"},
		MaxRunTime: 30,
	}

	td := testTask(t)

	taskID := scheduleDockerTask(t, td, payload)
	ensureResolution(t, taskID, "failed", "failed")
}

func TestInvalidCommand(t *testing.T) {
	defer setup(t)()

	payload := dockerworker.DockerWorkerPayload{
		Command:    []string{"/bin/invalid"},
		MaxRunTime: 30,
	}

	td := testTask(t)

	taskID := scheduleDockerTask(t, td, payload)
	ensureResolution(t, taskID, "failed", "failed")
}
