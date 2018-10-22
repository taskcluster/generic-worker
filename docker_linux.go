package main

import (
	"github.com/taskcluster/taskcluster-base-go/scopes"
)

type DockerFeature struct {
}

func (feature *DockerFeature) Name() string {
	return "Docker"
}

func (feature *DockerFeature) Initialise() error {
	return nil
}

func (feature *DockerFeature) PersistState() error {
	return nil
}

func (feature *DockerFeature) IsEnabled(task *TaskRun) bool {
	return true
}

type DockerTask struct {
	task *TaskRun
}

func (feature *DockerFeature) NewTaskFeature(task *TaskRun) TaskFeature {
	return &DockerTask{
		task: task,
	}
}

func (l *DockerTask) RequiredScopes() scopes.Required {
	return scopes.Required{}
}

func (l *DockerTask) Start() *CommandExecutionError {
	return nil
}

func (l *DockerTask) Stop() *CommandExecutionError {
	return nil
}
