// +build docker

package main

import (
	"encoding/json"
	"fmt"

	"github.com/taskcluster/generic-worker/dockerworker"
)

func dockerWorkerToGenericWorkerPayload(dw *dockerworker.DockerWorkerPayload) *GenericWorkerPayload {
	artifacts := make([]Artifact, 0, len(dw.Artifacts))
	for name, artifact := range dw.Artifacts {
		artifacts = append(artifacts, Artifact{
			Expires: artifact.Expires,
			Name:    name,
			Path:    artifact.Path,
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
	}
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

	task.PlatformData.Image = payload.Image
	task.Payload = *dockerWorkerToGenericWorkerPayload(&payload)

	return nil
}
