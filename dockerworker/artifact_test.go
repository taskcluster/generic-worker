// +build docker

package dockerworker

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
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

func runCommand(t *testing.T, d *DockerWorker, command []string) *docker.Container {
	img, err := d.LoadImage("ubuntu:14.04")
	require.NoError(t, err)

	container, err := d.CreateContainer(
		[]string{},
		img,
		command,
		false,
	)
	require.NoError(t, err)

	exitCode, _, err := d.RunContainer(container)
	require.NoError(t, err)
	require.Equal(t, exitCode, 0)

	return container
}

func TestExtractFile(t *testing.T) {
	const str = "But not for Suchong"

	d := NewTestDockerWorker(t)

	container := runCommand(t, d, []string{"/bin/bash", "-c", fmt.Sprintf("echo -n %s > /test.txt", str)})
	defer d.RemoveContainer(container)

	destdir := os.TempDir()
	require.NoError(t, d.ExtractArtifact(container, "/test.txt", destdir))
	artifact := filepath.Join(destdir, "test.txt")
	content, err := ioutil.ReadFile(artifact)
	require.NoError(t, err)

	defer os.Remove(artifact)
	require.Equal(t, str, string(content))
}

func buildCommand(filesystem map[string]interface{}) (cmd []string) {
	echo := func(file, content string) string {
		return fmt.Sprintf("echo -n %s > %s", content, file)
	}

	mkdir := func(dir string) string {
		return "mkdir -p " + dir
	}

	cd := func(dir string) string {
		return "cd " + dir
	}

	for key, val := range filesystem {
		switch v := val.(type) {
		case map[string]interface{}:
			cmd = append(cmd, mkdir(key), cd(key))
			cmd = append(cmd, buildCommand(v)...)
			cmd = append(cmd, cd(".."))
		case string:
			cmd = append(cmd, echo(key, v))
		}
	}

	return
}

func checkArtifactDir(t *testing.T, filesystem map[string]interface{}, destdir string) {
	for key, val := range filesystem {
		switch v := val.(type) {
		case map[string]interface{}:
			d := filepath.Join(destdir, key)
			_, err := os.Stat(d)
			require.NoError(t, err)
			checkArtifactDir(t, v, d)
		case string:
			content, err := ioutil.ReadFile(filepath.Join(destdir, key))
			require.NoError(t, err)
			require.Equal(t, string(content), v)
		}
	}
}

func TestExtractFolder(t *testing.T) {
	filesystem := map[string]interface{}{
		"testdir": map[string]interface{}{
			"file1.txt": "File 1 content",
			"testsubdir": map[string]interface{}{
				"test.txt": "This is a test file",
			},
			"testsubdir2": map[string]interface{}{
				"test2.txt": "This is another test file",
			},
		},
	}

	cmd := []string{"cd /"}
	cmd = append(cmd, buildCommand(filesystem)...)

	d := NewTestDockerWorker(t)
	destdir := os.TempDir()

	container := runCommand(t, d, []string{"/bin/bash", "-c", strings.Join(cmd, "; ")})
	defer d.RemoveContainer(container)

	defer os.RemoveAll(filepath.Join(destdir, "/testdir"))
	require.NoError(t, d.ExtractArtifact(container, "/testdir", destdir))

	checkArtifactDir(t, filesystem, destdir)
}
