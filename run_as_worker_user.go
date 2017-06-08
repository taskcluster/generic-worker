package main

import (
	"github.com/taskcluster/taskcluster-base-go/scopes"
)

// RunAsWorkerUser is a feature that is useful mostly for the CI - it allows
// the commands in a task to be executed as the user running the
// generic-worker, rather than running as a task user. On Windows this means
// executing in the context of a Windows service running as LocalSystem.
type RunAsWorkerUser struct {
}

type RunAsWorkerUserTaskFeature struct {
	task *TaskRun
}

func (feature *RunAsWorkerUser) Name() string {
	return "Run as Worker User"
}

func (feature *RunAsWorkerUser) PersistState() error {
	return nil
}

func (feature *RunAsWorkerUser) Initialise() error {
	return nil
}

func (feature *RunAsWorkerUser) IsEnabled(fl EnabledFeatures) bool {
	return fl.RunAsWorkerUser
}

func (feature *RunAsWorkerUser) NewTaskFeature(task *TaskRun) TaskFeature {
	return &RunAsWorkerUserTaskFeature{
		task: task,
	}
}

func (cot *RunAsWorkerUserTaskFeature) RequiredScopes() scopes.Required {
	// let's not require any scopes, as I see no reason to control access to this feature
	return scopes.Required{{"generic-worker:run-as-worker-user:" + config.WorkerType}}
}

func (cot *RunAsWorkerUserTaskFeature) Start() *CommandExecutionError {
	task.Log(" ***** Running task commands as worker user! *****")
	return nil
}

func (cot *RunAsWorkerUserTaskFeature) Stop() *CommandExecutionError {
	return nil
}
