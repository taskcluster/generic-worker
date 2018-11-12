// +build docker

package dockerworker

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taskcluster/httpbackoff"
)

const (
	// imageTaskID is the task containing the sample artifacts images for test
	imageTaskID   = "Xx0aPfyOTU2o_0FZnr_AJg"
	testNamespace = "garbage.docker-worker-tests.docker-images"
	exampleImage  = "coreos/example:1.0.0"
)

func loadArtifactImageTest(t *testing.T, taskID, artifactName string) {
	d := NewTestDockerWorker(t)

	imageName := MakeImageName(imageTaskID, "", artifactName)

	img, err := d.LoadArtifactImage(imageTaskID, "", artifactName)
	require.NoError(t, err)
	defer d.Client.RemoveImage(img.ID)

	loadedImage, err := d.Client.InspectImage(imageName)
	require.NoError(t, err)
	require.Equal(t, loadedImage.ID, img.ID)
}

func TestLoadArtifactImageTar(t *testing.T) {
	loadArtifactImageTest(t, imageTaskID, "public/image.tar")
}

func TestLoadArtifactImageZst(t *testing.T) {
	loadArtifactImageTest(t, imageTaskID, "public/image.tar.zst")
}

func TestLoadArtifactImageLz4(t *testing.T) {
	loadArtifactImageTest(t, imageTaskID, "public/image.tar.lz4")
}

func TestLoadIndexedImage(t *testing.T) {
	d := NewTestDockerWorker(t)

	img, err := d.LoadIndexedImage(testNamespace, "public/image.tar")
	require.NoError(t, err)

	require.NoError(t, d.Client.RemoveImage(img.ID))
}

func TestLoadImage(t *testing.T) {
	d := NewTestDockerWorker(t)

	img, err := d.LoadImage(exampleImage)
	require.NoError(t, err)
	defer d.Client.RemoveImage(img.ID)

	img2, err := d.Client.InspectImage(exampleImage)
	require.NoError(t, err)
	require.Equal(t, img.ID, img2.ID)
}

func TestImageWithManifest(t *testing.T) {
	const imageURL = "https://s3-us-west-2.amazonaws.com/docker-worker-manifest-test/image.tar.zst"
	const artifactName = "public/image.tar.zst"

	d := NewTestDockerWorker(t)

	resp, _, err := httpbackoff.Get(imageURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	taskID, err := createDummyTask(d)
	require.NoError(t, err)
	defer d.Queue.ReportCompleted(taskID, "0")

	require.NoError(t, uploadArtifact(d, taskID, "0", artifactName, resp.Body))

	loadArtifactImageTest(t, imageTaskID, "public/image.tar.zst")
}
