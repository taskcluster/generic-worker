// +build docker

package dockerworker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

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
	d.TaskLogger.Printf("Running \"%s\"", strings.Join(command, " "))
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

// RunContainers runs the passed container and extracts its logs
func (d *DockerWorker) RunContainer(container *docker.Container) (exitCode int, duration time.Duration, err error) {
	if err = d.Client.StartContainerWithContext(container.ID, nil, d.Context); err != nil {
		err = fmt.Errorf("Error starting container: %v", err)
		return
	}

	started := time.Now()
	if exitCode, err = d.Client.WaitContainer(container.ID); err != nil {
		err = fmt.Errorf("Error wating for container to finish: %v", err)
		return
	}

	err = d.Client.Logs(docker.LogsOptions{
		Context:      d.Context,
		Container:    container.ID,
		OutputStream: d.LivelogWriter,
		ErrorStream:  d.LivelogWriter,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
		Timestamps:   true,
	})

	if err != nil {
		err = fmt.Errorf("Error pulling container logs: %v", err)
		return
	}

	duration = time.Now().Sub(started)
	return
}

// RemoveContainer removes the given container from the system
func (d *DockerWorker) RemoveContainer(container *docker.Container) error {
	return d.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:      container.ID,
		Force:   true,
		Context: d.Context,
	})
}
