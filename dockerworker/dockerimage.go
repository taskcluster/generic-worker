// +build docker

package dockerworker

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/DataDog/zstd"
	"github.com/cenkalti/backoff"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pierrec/lz4"
	"github.com/taskcluster/taskcluster-client-go/tcindex"
)

func MakeImageName(taskID, runID, artifactName string) string {
	return strings.ToLower(base64.RawStdEncoding.EncodeToString([]byte(taskID + runID + artifactName)))
}

// Ensure image pulls the given image if it doesn't exist
func EnsureImage(ctx context.Context, cli *docker.Client, imageName string, log io.Writer) (img *docker.Image, err error) {
	if img, err = cli.InspectImage(imageName); err == nil {
		return
	}

	err = backoff.Retry(func() error {
		return cli.PullImage(docker.PullImageOptions{
			Repository:   imageName,
			OutputStream: log,
			Context:      ctx,
		}, docker.AuthConfiguration{})
	}, backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), ctx), 3))

	if err != nil {
		return
	}

	return cli.InspectImage(imageName)
}

// LoadArtifactImage loads an image from a task artifact
func (d *DockerWorker) LoadArtifactImage(taskID, runID, name string) (img *docker.Image, err error) {
	if runID != "" {
		d.TaskLogger.Printf("Loading image %s/%s/%s", taskID, runID, name)
	} else {
		d.TaskLogger.Printf("Loading image %s/%s", taskID, name)
	}

	imageName := MakeImageName(taskID, runID, name)

	// If we already downloaded the image, just return it
	img, err = d.Client.InspectImage(imageName)
	if err == nil {
		return
	}

	// The downloaded image is written to a file because it might be
	// a huge image
	f, err := ioutil.TempFile(os.TempDir(), "image")
	if err != nil {
		return
	}

	defer os.Remove(f.Name())
	defer f.Close()

	if err = d.DownloadArtifact(taskID, runID, name, f); err != nil {
		return
	}

	// rewind the file pointer to read the downloaded file
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return
	}

	var r io.Reader

	// We accept the image in three formats: .tar, .tar.lz4 and .tar.zst
	switch {
	case strings.HasSuffix(name, ".zst"):
		rc := zstd.NewReader(f)
		defer rc.Close()
		r = rc
	case strings.HasSuffix(name, ".lz4"):
		r = lz4.NewReader(f)
	case strings.HasSuffix(name, ".tar"):
		r = f
	default:
		err = fmt.Errorf("Not supported format for image artifact %s", name)
		return
	}

	// We need to rename the image name to avoid cache poisoning.
	// Again we store it in a temporary file to avoid exhausting
	// the memory
	t, err := ioutil.TempFile(os.TempDir(), "image.tar")
	if err != nil {
		return
	}

	defer os.Remove(t.Name())
	defer t.Close()

	// As we cache images, we need to give it a unique name,
	// other wise an attacker can use a cache poisoning attack
	// to make a task load malicious image. Ref: Bug 1389719
	err = renameDockerImageTarStream(imageName, r, t)
	if err != nil {
		return
	}

	if _, err = t.Seek(0, io.SeekStart); err != nil {
		return
	}

	d.Client.LoadImage(docker.LoadImageOptions{
		Context:      d.Context,
		InputStream:  t,
		OutputStream: d.LivelogWriter,
	})

	img, err = d.Client.InspectImage(imageName)

	return
}

func (d *DockerWorker) LoadIndexedImage(namespace, path string) (*docker.Image, error) {
	d.TaskLogger.Printf("Loading image %s:%s", namespace, path)

	index := tcindex.NewFromEnv()

	task, err := index.FindTask(namespace)
	if err != nil {
		return nil, err
	}

	return d.LoadArtifactImage(task.TaskID, "", path)
}

func (d *DockerWorker) LoadImage(name string) (*docker.Image, error) {
	d.TaskLogger.Printf("Loading image %s", name)
	return EnsureImage(d.Context, d.Client, name, d.LivelogWriter)
}
