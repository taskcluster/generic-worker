package main

import (
	"net"
	"path/filepath"
	"time"

	"github.com/taskcluster/generic-worker/fileutil"
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
	info *RDPInfo
}

type RDPInfo struct {
	Host     net.IP `json:"host"`
	Port     uint16 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
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
	l.createRDPArtifact()
	return l.uploadRDPArtifact()
}

func (l *RDPTask) Stop() *CommandExecutionError {
	time.Sleep(time.Hour * 12)
	return nil
}

func (l *RDPTask) createRDPArtifact() {
	autoLogonUser, autoLogonPassword := AutoLogonCredentials()
	l.info = &RDPInfo{
		Host:     config.PublicIP,
		Port:     3389,
		Username: autoLogonUser,
		Password: autoLogonPassword,
	}
	err := fileutil.WriteToFileAsJSON(l.info, rdpArtifactPath)
	// if we can't write this, something seriously wrong, so cause worker to
	// report an internal-error to sentry and crash!
	if err != nil {
		panic(err)
	}
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
