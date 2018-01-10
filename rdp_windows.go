package main

import (
	"path/filepath"
	"time"

	"github.com/taskcluster/taskcluster-base-go/scopes"
	tcclient "github.com/taskcluster/taskcluster-client-go"
)

var (
	rdpArtifactPath = filepath.Join("generic-worker", "rdp.json")
)

type RDPFeature struct {
}

func (feature *RDPFeature) Name() string {
	return "RDP"
}

func (feature *RDPFeature) Initialise() error {
	return nil
}

func (feature *RDPFeature) PersistState() error {
	return nil
}

// RDP is only enabled when task.payload.rdpInfo is set
func (feature *RDPFeature) IsEnabled(task *TaskRun) bool {
	return task.Payload.RdpInfo != ""
}

type RDPTask struct {
	task *TaskRun
}

func (feature *RDPFeature) NewTaskFeature(task *TaskRun) TaskFeature {
	return &RDPTask{
		task: task,
	}
}

func (l *RDPTask) RequiredScopes() scopes.Required {
	// let's not require any scopes, as I see no reason to control access to this feature
	return scopes.Required{
		{
			"generic-worker:allow-rdp:" + l.task.Definition.ProvisionerID + "/" + l.task.Definition.WorkerType,
		},
	}
}

func (l *RDPTask) ReservedArtifacts() []string {
	return []string{
		l.task.Payload.RdpInfo,
	}
}

func (l *RDPTask) Start() *CommandExecutionError {
	return l.uploadRDPArtifact()
}

func (l *RDPTask) Stop() *CommandExecutionError {
	time.Sleep(time.Hour * 12)
	return nil
}

func (l *RDPTask) uploadRDPArtifact() *CommandExecutionError {
	return l.task.uploadArtifact(
		&S3Artifact{
			BaseArtifact: &BaseArtifact{
				Name: l.task.Payload.RdpInfo,
				// RDP info expires one day after task
				Expires:  tcclient.Time(time.Now().Add(time.Hour * 24)),
				MimeType: "application/json",
			},
			Path: rdpArtifactPath,
		},
	)
}
