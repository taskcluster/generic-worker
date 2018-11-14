// +build docker

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/taskcluster/generic-worker/dockerworker"
)

func (task *TaskRun) dockerWorkerToGenericWorkerPayload(dw *dockerworker.DockerWorkerPayload) (*GenericWorkerPayload, error) {
	artifacts := make([]Artifact, 0, len(dw.Artifacts))
	for name, artifact := range dw.Artifacts {
		// Artifact paths in generic-worker are relative to the task directory
		relativePath := artifact.Path[1:]

		// docker-worker specifies absolute artifact paths, while generic-worker artifact
		// paths are relative to the task directory. We download the container artifact to
		// the same directory tree in the generic-worker root by the task dir.
		artifactDir := filepath.Dir(filepath.Join(taskContext.TaskDir, relativePath))
		if err := os.MkdirAll(artifactDir, 0700); err != nil && err != os.ErrExist {
			return nil, fmt.Errorf("Error creating artifact directory '%s': %v", artifactDir, err)
		}

		artifacts = append(artifacts, Artifact{
			Expires: artifact.Expires,
			Name:    name,
			Path:    relativePath,
			Type:    artifact.Type,
		})
	}

	return &GenericWorkerPayload{
		Artifacts:     artifacts,
		Command:       [][]string{dw.Command},
		Env:           dw.Env,
		MaxRunTime:    int64(dw.MaxRunTime),
		SupersederURL: dw.SupersederURL,
		Features:      FeatureFlags{},
		Mounts:        []json.RawMessage{},
		OnExitStatus:  ExitCodeHandling{},
		OSGroups:      []string{},
	}, nil
}

func (task *TaskRun) validatePayload() *CommandExecutionError {
	result, err := dockerworker.ValidatePayload(task.Definition.Payload)
	if err != nil {
		return MalformedPayloadError(err)
	}

	if !result.Valid() {
		task.Error("TASK FAIL since the task payload is invalid. See errors:")
		for _, desc := range result.Errors() {
			task.Errorf("- %s", desc)
		}
		return MalformedPayloadError(fmt.Errorf("Validation of payload failed for task %v", task.TaskID))
	}

	var payload dockerworker.DockerWorkerPayload
	if err = json.Unmarshal(task.Definition.Payload, &payload); err != nil {
		return MalformedPayloadError(err)
	}

	for name, artifact := range payload.Artifacts {
		if !filepath.IsAbs(artifact.Path) {
			return MalformedPayloadError(
				fmt.Errorf("The artifact paths must be absolute, but the path of '%s' isn't", name))
		}
	}

	task.PlatformData.Image = payload.Image
	p, err := task.dockerWorkerToGenericWorkerPayload(&payload)
	if err != nil {
		return executionError(internalError, errored, err)
	}
	task.Payload = *p

	return nil
}
