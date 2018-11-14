// +build docker

package main

import (
	"fmt"
	"testing"

	"github.com/taskcluster/generic-worker/dockerworker"
)

func TestFileArtifact(t *testing.T) {
	defer setup(t)()

	const message = "I am here to ask you a question"

	td := testTask(t)

	payload := dockerworker.DockerWorkerPayload{
		Command:    []string{"/bin/bash", "-c", fmt.Sprintf("/bin/echo -n %s > /home/test.txt", message)},
		MaxRunTime: 30,
		Artifacts: map[string]dockerworker.Artifact{
			"public/test.txt": dockerworker.Artifact{
				Expires: td.Expires,
				Path:    "/home/test.txt",
				Type:    "file",
			},
		},
	}

	expectedArtifacts := ExpectedArtifacts{
		"public/test.txt": {
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

func TestFolderArtifact(t *testing.T) {
	defer setup(t)()

	td := testTask(t)

	payload := dockerworker.DockerWorkerPayload{
		Command: []string{"/bin/bash", "-c", `
			mkdir -p /home/testdir/subdir;
			echo -n This is a test file > /home/testdir/test.txt;
			echo -n This is another test file > /home/testdir/subdir/anothertest.txt
		`},
		MaxRunTime: 30,
		Artifacts: map[string]dockerworker.Artifact{
			"public/testdir": dockerworker.Artifact{
				Expires: td.Expires,
				Path:    "/home/testdir",
				Type:    "directory",
			},
		},
	}

	expectedArtifacts := ExpectedArtifacts{
		"public/testdir/test.txt": {
			Extracts:        []string{"This is a test file"},
			ContentType:     "text/plain; charset=utf-8",
			ContentEncoding: "gzip",
			Expires:         td.Expires,
		},
		"public/testdir/subdir/anothertest.txt": {
			Extracts:        []string{"This is another test file"},
			ContentType:     "text/plain; charset=utf-8",
			ContentEncoding: "gzip",
			Expires:         td.Expires,
		},
	}

	taskID := scheduleDockerTask(t, td, payload)
	ensureResolution(t, taskID, "completed", "completed")

	expectedArtifacts.Validate(t, taskID, 0)
}

func TestArtifactNotFound(t *testing.T) {
	defer setup(t)()

	td := testTask(t)

	payload := dockerworker.DockerWorkerPayload{
		Command:    []string{"/bin/echo", "No artifact"},
		MaxRunTime: 30,
		Artifacts: map[string]dockerworker.Artifact{
			"public/test.txt": dockerworker.Artifact{
				Expires: td.Expires,
				Path:    "/home/test.txt",
				Type:    "file",
			},
		},
	}

	taskID := scheduleDockerTask(t, td, payload)
	ensureResolution(t, taskID, "failed", "failed")
}
