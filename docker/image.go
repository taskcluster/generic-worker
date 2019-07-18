package docker

import (
	"encoding/json"
	"fmt"
	"log"
)

var ImageNotSupportedError = fmt.Errorf("unsupported object found for task image")

type possibleImage struct {
	// common
	Type string `json:"type"`
	// if type == "docker-image"
	Name string `json:"name"`
	// common to "indexed-image","task-image"
	Path string `json:"Path"`
	// if type == "indexed-image"
	Namespace string `json:"namespace"`
	// if type == "task-image"
	TaskID string `json:"taskId"`
}

func ImageFromJSON(image json.RawMessage) (string, error) {
	// if image is invalid json at this point
	// we have a string docker image name
	if !json.Valid(image) {
		return string(image), nil
	}

	// handle object possibilities
	p := possibleImage{}
	err := json.Unmarshal(image, &p)
	if err != nil {
		return "", err
	}
	if p.Type == "docker-image" {
		log.Printf("got %#v, so it's a docker-image", p)
		return p.Name, nil
	}
	if p.Type == "indexed-image" {
		log.Printf("got %#v, so it's a indexed-image", p)
		return "", ImageNotSupportedError
	}
	if p.Type == "task-image" {
		log.Printf("got %#v, so it's a task-image", p)
		return "", ImageNotSupportedError
	}
	return "", fmt.Errorf("could not parse image:%s", string(image))
}
