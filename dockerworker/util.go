// +build docker

package dockerworker

import (
	"context"
	"os"
	"testing"

	"github.com/taskcluster/taskcluster-client-go/tcqueue"
)

const (
	TestProvisionerID = "null-provisioner"
	TestWorkerType    = "docker-worker"
	TestWorkerGroup   = "docker-worker"
	TestWorkerID      = "docker-worker"
	TestSchedulerID   = "docker-worker-tests"
)

// NewTestDockerWorker returns a new DockerWorker object for tests
func NewTestDockerWorker(t *testing.T) *DockerWorker {
	return New(context.Background(), tcqueue.NewFromEnv(), "testtaskid", os.Stdout)
}
