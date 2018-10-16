// +build docker

package process

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/mattetti/filebuffer"
	"github.com/stretchr/testify/require"
	"github.com/taskcluster/generic-worker/dockerworker"
)

func TestProcess(t *testing.T) {
	const message = "Would you kindly?"
	const imageName = "ubuntu:14.04"

	image, err := json.Marshal(imageName)
	require.NoError(t, err)

	d := dockerworker.NewTestDockerWorker(t)
	if _, err = d.Client.InspectImage(imageName); err == nil {
		require.NoError(t, d.Client.RemoveImageExtended(imageName, docker.RemoveImageOptions{
			Force:   true,
			Context: d.Context,
		}))
	}
	buff := filebuffer.New([]byte{})
	d.LivelogWriter = buff

	c, err := NewCommand(
		d,
		[]string{"/bin/bash", "-c", "echo $MESSAGE"},
		image,
		"/",
		[]string{fmt.Sprintf("MESSAGE='%s'", message)},
	)
	require.NoError(t, err)

	r := c.Execute()
	require.NoError(t, r.SystemError)
	require.Equal(t, r.ExitCode(), 0)

	_, err = buff.Seek(0, io.SeekStart)
	require.NoError(t, err)

	require.True(t, strings.Contains(buff.String(), message))
}
