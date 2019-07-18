package docker

import (
	"fmt"
	"testing"
)

func TestHandleDockerImage(t *testing.T) {
	var image, expected string
	var input []byte
	var err error

	// handle image is a string docker image
	expected = "ubuntu:18.04"
	input = []byte(fmt.Sprintf("%v", expected))
	image, err = ImageFromJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if image != expected {
		t.Fatalf("Expected %s, got %#v", expected, image)
	}

	// handle type docker-image
	expected = "ubuntu:18.04"
	input = []byte(fmt.Sprintf(`{
		"name": "%s",
		"type": "docker-image"
	}`, expected))
	image, err = ImageFromJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if image != expected {
		t.Fatalf("Expected: %s, got: %#v", expected, image)
	}

	// handle type task-image
	expected = "AFelkflgjalall3402"
	input = []byte(fmt.Sprintf(`{
		"path": "public/image.tar.zst",
		"taskId": "%s",
		"type": "task-image"
	}`, expected))
	image, err = ImageFromJSON(input)
	// not implemented yet
	if err != ImageNotSupportedError {
		t.Fatalf("Expected: %#v, got: %v", ImageNotSupportedError, err)
	}

	// handle type indexed-image
	expected = "flarb"
	input = []byte(fmt.Sprintf(`{
		"path": "public/image.tar.zst",
		"namespace": "%s",
		"type": "indexed-image"
	}`, expected))
	image, err = ImageFromJSON(input)
	// not implemented yet
	if err != ImageNotSupportedError {
		t.Fatalf("Expected: %#v, got: %v", ImageNotSupportedError, err)
	}
}
