// +build docker

package dockerworker

import (
	"io"
	"testing"

	"github.com/mattetti/filebuffer"
	"github.com/stretchr/testify/require"
)

func TestArtifact(t *testing.T) {
	d := NewTestDockerWorker(t)

	content := "This is a dummy artifact"

	taskID, err := createDummyTask(d)
	require.NoError(t, err)
	defer d.Queue.ReportCompleted(taskID, "0")

	require.NoError(t, createDummyArtifact(d, taskID, "public/test", content))

	// Test Download Latest
	buff := filebuffer.New([]byte{})
	require.NoError(t, d.DownloadArtifact(taskID, "", "public/test", buff))

	_, err = buff.Seek(0, io.SeekStart)
	require.NoError(t, err)
	require.Equal(t, content, buff.String())

	// Test Download specific runID
	buff = filebuffer.New([]byte{})
	require.NoError(t, d.DownloadArtifact(taskID, "0", "public/test", buff))

	_, err = buff.Seek(0, io.SeekStart)
	require.NoError(t, err)
	require.Equal(t, content, buff.String())
}
