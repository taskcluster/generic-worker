package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"

	"github.com/taskcluster/generic-worker/lib/fileutil"
	"github.com/taskcluster/taskcluster-base-go/scopes"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
)

const (
	// ChainOfTrustKeyNotSecureMessage contains message to log when chain of
	// trust key is discovered at runtime not to be secure
	ChainOfTrustKeyNotSecureMessage = "Was expecting attempt to read private chain of trust key as task user to fail - however, it did not!"
)

var (
	certifiedLogPath = filepath.Join("generic-worker", "certified.log")
	certifiedLogName = "public/logs/certified.log"
	signedCertPath   = filepath.Join("generic-worker", "chainOfTrust.json.asc")
	signedCertName   = "public/chainOfTrust.json.asc"
)

type ChainOfTrustFeature struct {
	PrivateKey *packet.PrivateKey
}

type ArtifactHash struct {
	SHA256 string `json:"sha256"`
}

type CoTEnvironment struct {
	PublicIPAddress  string `json:"publicIpAddress"`
	PrivateIPAddress string `json:"privateIpAddress"`
	InstanceID       string `json:"instanceId"`
	InstanceType     string `json:"instanceType"`
	Region           string `json:"region"`
}

type ChainOfTrustData struct {
	Version     int                            `json:"chainOfTrustVersion"`
	Artifacts   map[string]ArtifactHash        `json:"artifacts"`
	Task        tcqueue.TaskDefinitionResponse `json:"task"`
	TaskID      string                         `json:"taskId"`
	RunID       uint                           `json:"runId"`
	WorkerGroup string                         `json:"workerGroup"`
	WorkerID    string                         `json:"workerId"`
	Environment CoTEnvironment                 `json:"environment"`
}

type ChainOfTrustTaskFeature struct {
	task    *TaskRun
	privKey *packet.PrivateKey
}

func (feature *ChainOfTrustFeature) Name() string {
	return "Chain of Trust"
}

func (feature *ChainOfTrustFeature) PersistState() error {
	return nil
}

func (feature *ChainOfTrustFeature) Initialise() (err error) {
	feature.PrivateKey, err = readPrivateKey()
	if err != nil {
		return
	}

	// platform-specific mechanism to lock down file permissions
	// of private signing key
	err = secureSigningKey()
	return
}

func readPrivateKey() (privateKey *packet.PrivateKey, err error) {
	var privKeyFile *os.File
	privKeyFile, err = os.Open(config.SigningKeyLocation)
	if err != nil {
		log.Printf("FATAL: Was not able to open chain of trust signing key file '%v'.", config.SigningKeyLocation)
		log.Printf("The chain of trust signing key file location is configured in file '%v' in property 'signingKeyLocation'.", configFile)
		return
	}
	defer privKeyFile.Close()
	var entityList openpgp.EntityList
	entityList, err = openpgp.ReadArmoredKeyRing(privKeyFile)
	if err != nil {
		return
	}
	privateKey = entityList[0].PrivateKey
	return
}

func (feature *ChainOfTrustFeature) IsEnabled(task *TaskRun) bool {
	return task.Payload.Features.ChainOfTrust
}

func (feature *ChainOfTrustFeature) NewTaskFeature(task *TaskRun) TaskFeature {
	return &ChainOfTrustTaskFeature{
		task:    task,
		privKey: feature.PrivateKey,
	}
}

func (feature *ChainOfTrustTaskFeature) ReservedArtifacts() []string {
	return []string{
		signedCertName,
		certifiedLogName,
	}
}

func (feature *ChainOfTrustTaskFeature) RequiredScopes() scopes.Required {
	// let's not require any scopes, as I see no reason to control access to this feature
	return scopes.Required{}
}

func (feature *ChainOfTrustTaskFeature) Start() *CommandExecutionError {
	// Return an error if the task user can read the private key file.
	// We shouldn't be able to read the private key, if we can let's raise
	// MalformedPayloadError, as it could be a problem with the task definition
	// (for example, enabling chainOfTrust on a worker type that has
	// runTasksAsCurrentUser enabled).
	err := feature.ensureTaskUserCantReadPrivateCotKey()
	if err != nil {
		return MalformedPayloadError(err)
	}
	return nil
}

func (feature *ChainOfTrustTaskFeature) Stop(err *ExecutionErrors) {
	logFile := filepath.Join(taskContext.TaskDir, logPath)
	certifiedLogFile := filepath.Join(taskContext.TaskDir, certifiedLogPath)
	signedCert := filepath.Join(taskContext.TaskDir, signedCertPath)
	copyErr := copyFileContents(logFile, certifiedLogFile)
	if copyErr != nil {
		panic(copyErr)
	}
	err.add(feature.task.uploadLog(certifiedLogName, certifiedLogPath))
	artifactHashes := map[string]ArtifactHash{}
	for _, artifact := range feature.task.Artifacts {
		switch a := artifact.(type) {
		case *S3Artifact:
			// make sure SHA256 is calculated
			file := filepath.Join(taskContext.TaskDir, a.Path)
			hash, hashErr := fileutil.CalculateSHA256(file)
			if hashErr != nil {
				panic(hashErr)
			}
			artifactHashes[a.Name] = ArtifactHash{
				SHA256: hash,
			}
		}
	}

	cotCert := &ChainOfTrustData{
		Version:     1,
		Artifacts:   artifactHashes,
		Task:        feature.task.Definition,
		TaskID:      feature.task.TaskID,
		RunID:       feature.task.RunID,
		WorkerGroup: config.WorkerGroup,
		WorkerID:    config.WorkerID,
		Environment: CoTEnvironment{
			PublicIPAddress:  config.PublicIP.String(),
			PrivateIPAddress: config.PrivateIP.String(),
			InstanceID:       config.InstanceID,
			InstanceType:     config.InstanceType,
			Region:           config.Region,
		},
	}

	certBytes, e := json.MarshalIndent(cotCert, "", "  ")
	if e != nil {
		panic(e)
	}
	// separate signature from json with a new line
	certBytes = append(certBytes, '\n')

	in := bytes.NewBuffer(certBytes)
	out, e := os.Create(signedCert)
	if e != nil {
		panic(e)
	}
	defer out.Close()

	w, e := clearsign.Encode(out, feature.privKey, nil)
	if e != nil {
		panic(e)
	}
	_, e = io.Copy(w, in)
	if e != nil {
		panic(e)
	}
	w.Close()
	out.Write([]byte{'\n'})
	out.Close()
	err.add(feature.task.uploadLog(signedCertName, signedCertPath))
}
