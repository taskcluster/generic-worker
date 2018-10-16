// +build docker

// This module contains utility functions for package tests

package dockerworker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/taskcluster/slugid-go/slugid"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
	artifact "github.com/taskcluster/taskcluster-lib-artifact-go"
)

// scheduleReclaim run a goroutine that will perform reclaims in the background.
func scheduleReclaim(d *DockerWorker, claim *tcqueue.TaskClaimResponse) {
	takenUntil := claim.TakenUntil

	go func() {
		for {
			select {
			case <-time.After(time.Time(takenUntil).Sub(time.Now())):
				reclaim, err := d.Queue.ReclaimTask(claim.Status.TaskID, fmt.Sprint(claim.RunID))
				if err != nil {
					d.TaskLogger.Printf("Exiting reclaim loop: %v", err)
					return
				}
				takenUntil = reclaim.TakenUntil
			case <-d.Context.Done():
				return
			}
		}
	}()
}

// uploadArtifact uploads a new artifact to the task
func uploadArtifact(d *DockerWorker, taskID, runID, name string, in io.Reader) (ret error) {
	client := artifact.New(d.Queue)

	tmp, ret := ioutil.TempFile(os.TempDir(), slugid.Nice())
	if ret != nil {
		return
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	// Use a temporary file because we need a ReadSeeker interface
	if _, ret = io.Copy(tmp, in); ret != nil {
		return
	}

	backoffErr := backoff.Retry(func() error {
		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			ret = err
			return nil
		}

		f, err := ioutil.TempFile(os.TempDir(), "gw")
		if err != nil {
			return err
		}

		defer os.Remove(f.Name())
		defer f.Close()

		return client.Upload(taskID, runID, name, tmp, f, false, false)
	}, backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), d.Context), 3))

	if ret == nil {
		ret = backoffErr
	}

	return
}

func createDummyTask(d *DockerWorker) (taskID string, err error) {
	task := tcqueue.TaskDefinitionRequest{
		Created:  tcclient.Time(time.Now()),
		Deadline: tcclient.Time(time.Now().Add(10 * time.Minute)),
		Expires:  tcclient.Time(time.Now().Add(48 * time.Hour)),
		Payload:  json.RawMessage("{}"),
		Metadata: tcqueue.TaskMetadata{
			Description: "generic-worker dummy task",
			Name:        "generic-worker dummy task",
			Owner:       "wcosta@mozilla.com",
			Source:      "https://www.mozilla.org",
		},
		Priority:      "lowest",
		ProvisionerID: TestProvisionerID,
		WorkerType:    TestWorkerType,
		SchedulerID:   TestSchedulerID,
	}

	taskID = slugid.Nice()

	if _, err = d.Queue.CreateTask(taskID, &task); err != nil {
		return
	}

	fmt.Printf("Created task %s\n", taskID)

	claimResp, err := d.Queue.ClaimTask(taskID, "0", &tcqueue.TaskClaimRequest{
		WorkerGroup: TestWorkerGroup,
		WorkerID:    TestWorkerID,
	})

	if err != nil {
		return
	}

	scheduleReclaim(d, claimResp)

	return
}

func createDummyArtifact(d *DockerWorker, taskID, name, content string) error {
	return uploadArtifact(d, taskID, "0", name, strings.NewReader(content))
}
