// +build docker

package process

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/taskcluster/generic-worker/dockerworker"
)

type Result struct {
	SystemError error
	ExitError   error
	exitCode    int
	Duration    time.Duration
}

type Command struct {
	mutex            sync.RWMutex
	worker           *dockerworker.DockerWorker
	writer           io.Writer
	cmd              []string
	workingDirectory string
	image            json.RawMessage
	env              []string
}

func (c *Command) ensureImage() (img *docker.Image, err error) {
	var imageName string
	var imageArtifact dockerworker.DockerImageArtifact
	var indexedImage dockerworker.IndexedDockerImage

	if err = json.Unmarshal(c.image, &imageName); err == nil {
		img, err = c.worker.LoadImage(imageName)
	} else if err = json.Unmarshal(c.image, &imageArtifact); err == nil {
		img, err = c.worker.LoadArtifactImage(imageArtifact.TaskID, "", imageArtifact.Path)
	} else if err = json.Unmarshal(c.image, &indexedImage); err == nil {
		img, err = c.worker.LoadIndexedImage(indexedImage.Namespace, indexedImage.Path)
	} else {
		err = errors.New("Invalid image specification")
	}

	return
}

func (c *Command) DirectOutput(writer io.Writer) {
	c.writer = writer
}

func (c *Command) String() string {
	return fmt.Sprintf("%q", c.cmd)
}

func (c *Command) Execute() (r *Result) {
	r = &Result{}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	image, err := c.ensureImage()
	if err != nil {
		r.SystemError = fmt.Errorf("Error downloading image: %v", err)
		return
	}

	container, err := c.worker.CreateContainer(c.env, image, c.cmd, false)
	if err != nil {
		r.SystemError = fmt.Errorf("Error creating a new container: %v", err)
		return
	}

	if err = c.worker.Client.StartContainerWithContext(container.ID, nil, c.worker.Context); err != nil {
		r.ExitError = fmt.Errorf("Error starting container: %v", err)
		return
	}

	started := time.Now()
	if r.exitCode, err = c.worker.Client.WaitContainer(container.ID); err != nil {
		r.ExitError = fmt.Errorf("Error wating for container to finish: %v", err)
		return
	}

	err = c.worker.Client.Logs(docker.LogsOptions{
		Context:      c.worker.Context,
		Container:    container.ID,
		OutputStream: c.worker.LivelogWriter,
		ErrorStream:  c.worker.LivelogWriter,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
		Timestamps:   true,
	})
	if err != nil {
		r.SystemError = fmt.Errorf("Error pulling container logs: %v", err)
	}

	finished := time.Now()
	r.Duration = finished.Sub(started)

	return
}

func (r *Result) CrashCause() error {
	return r.SystemError
}

func (r *Result) Crashed() bool {
	return r.SystemError != nil
}

func (r *Result) FailureCause() error {
	if r.ExitError == nil {
		return fmt.Errorf("Exit code %v", r.exitCode)
	}

	return r.ExitError
}

func (r *Result) Failed() bool {
	return r.SystemError == nil && (r.exitCode != 0 || r.ExitError != nil)
}

func (r *Result) ExitCode() int {
	if r.SystemError != nil || r.ExitError != nil {
		return -2
	}

	return r.exitCode
}

func NewCommand(worker *dockerworker.DockerWorker, commandLine []string, image json.RawMessage, workingDirectory string, env []string) (*Command, error) {
	c := &Command{
		worker:           worker,
		writer:           os.Stdout,
		cmd:              commandLine,
		workingDirectory: workingDirectory,
		image:            image,
		env:              env,
	}
	return c, nil
}

func (c *Command) Kill() ([]byte, error) {
	return []byte{}, nil
}
