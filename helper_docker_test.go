// +build docker

package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taskcluster/generic-worker/dockerworker"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
)

func scheduleDockerTask(t *testing.T, td *tcqueue.TaskDefinitionRequest, payload dockerworker.DockerWorkerPayload) string {
	b, err := json.Marshal("ubuntu:14.04")
	require.NoError(t, err)

	imageJSON := json.RawMessage{}
	require.NoError(t, json.Unmarshal(b, &imageJSON))
	payload.Image = imageJSON

	b, err = json.Marshal(&payload)
	require.NoError(t, err)

	payloadJSON := json.RawMessage{}
	require.NoError(t, json.Unmarshal(b, &payloadJSON))
	td.Payload = payloadJSON

	return scheduleTask(t, td, GenericWorkerPayload{})
}
