// +build docker

package dockerworker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// imageTaskID is the task containing the sample artifacts images for test
	imageTaskID   = "Xx0aPfyOTU2o_0FZnr_AJg"
	testNamespace = "garbage.docker-worker-tests.docker-images"
	exampleImage  = "coreos/example:1.0.0"
)

func loadArtifactImageTest(t *testing.T, artifactName string) {
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
	loadArtifactImageTest(t, "public/image.tar")
}

func TestLoadArtifactImageZst(t *testing.T) {
	loadArtifactImageTest(t, "public/image.tar.zst")
}

func TestLoadArtifactImageLz4(t *testing.T) {
	loadArtifactImageTest(t, "public/image.tar.lz4")
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
