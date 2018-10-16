// +build docker

package dockerworker

import (
	"context"
	"encoding/json"
	"io"
	"log"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
	"github.com/xeipuuv/gojsonschema"
)

var cli *docker.Client

func init() {
	var err error
	if cli, err = docker.NewClientFromEnv(); err != nil {
		panic(err)
	}
}

type DockerWorker struct {
	LivelogWriter io.Writer
	Logger        *log.Logger
	TaskLogger    *log.Logger
	Context       context.Context
	Queue         *tcqueue.Queue
	Client        *docker.Client
}

func New(ctx context.Context, queue *tcqueue.Queue, taskID string, liveLogWriter io.Writer) *DockerWorker {
	return &DockerWorker{
		LivelogWriter: liveLogWriter,
		Logger:        NewLogger(taskID),
		TaskLogger:    NewTaskLogger(liveLogWriter),
		Context:       ctx,
		Queue:         queue,
		Client:        cli,
	}
}

// ValidatePayload validates the docker worker task payload
func ValidatePayload(payload json.RawMessage) (result *gojsonschema.Result, err error) {
	schemaLoader := gojsonschema.NewStringLoader(taskPayloadSchema())
	docLoader := gojsonschema.NewStringLoader(string(payload))
	return gojsonschema.Validate(schemaLoader, docLoader)
}

// CreateContainer creates a new docker container to run a task
func (d *DockerWorker) CreateContainer(env []string, image *docker.Image, command []string, privileged bool) (container *docker.Container, err error) {
	return d.Client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        image.ID,
			Cmd:          command,
			Hostname:     "",
			User:         "",
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    false,
			StdinOnce:    false,
			Env:          env,
		},
		HostConfig: &docker.HostConfig{
			Privileged: privileged,
			ShmSize:    1800000000,
			ExtraHosts: []string{
				"localhost.localdomain:127.0.0.1", // Bug 1488148
			},
		},
		Context: d.Context,
	})
}
