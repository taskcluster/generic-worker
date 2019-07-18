// +build docker

package main

import (
	"log"
	"os"

	"github.com/taskcluster/generic-worker/docker"
	"github.com/taskcluster/generic-worker/process"
)

const (
	engine = "docker"
)

func secureConfigFile() {
}

func MkdirAllTaskUser(dir string, perms os.FileMode) (err error) {
	return nil
}

func (task *TaskRun) generateCommand(index int) error {
	image, err := docker.ImageFromJSON(task.Payload.Image)
	if err != nil {
		image = "ubuntu"
		log.Printf("Error parsing Image provided in task payload: %v", err)
		log.Printf("Falling back to default image %q", image)
	}
	task.Commands[index], err = process.NewDockerCommand(task.Payload.Command[index], image, taskContext.TaskDir, task.EnvVars())
	if err != nil {
		return err
	}
	task.logMux.RLock()
	defer task.logMux.RUnlock()
	task.Commands[index].DirectOutput(task.logWriter)
	return nil
}
