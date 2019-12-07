package main

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/taskcluster/slugid-go/slugid"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
)

var (
	// all tests can share taskGroupId so we can view all test tasks in same
	// graph later for troubleshooting
	taskGroupID = slugid.Nice()
)

func validateArtifacts(
	t *testing.T,
	payloadArtifacts []Artifact,
	expected []TaskArtifact) {

	// to test, create a dummy task run with given artifacts
	// and then call Artifacts() method to see what
	// artifacts would get uploaded...
	tr := &TaskRun{
		Payload: GenericWorkerPayload{
			Artifacts: []Artifact{},
		},
		Definition: tcqueue.TaskDefinitionResponse{
			Expires: inAnHour,
		},
	}
	for i := range payloadArtifacts {
		tr.Payload.Artifacts = append(tr.Payload.Artifacts, payloadArtifacts[i])
	}
	artifacts := tr.PayloadArtifacts()

	if !reflect.DeepEqual(artifacts, expected) {
		t.Fatalf("Expected different artifacts to be generated...\nExpected:\n%q\nActual:\n%q", expected, artifacts)
	}
}

func TestFileArtifactWithNames(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{
			{
				Expires: inAnHour,
				Path:    "SampleArtifacts/_/X.txt",
				Type:    "file",
				Name:    "public/build/firefox.exe",
			},
		},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/build/firefox.exe",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            "SampleArtifacts/_/X.txt",
			},
		})
}

func TestFileArtifactWithContentType(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{
			{
				Expires:     inAnHour,
				Path:        "SampleArtifacts/_/X.txt",
				Type:        "file",
				Name:        "public/build/firefox.exe",
				ContentType: "application/octet-stream",
			},
		},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/build/firefox.exe",
					Expires: inAnHour,
				},
				ContentType:     "application/octet-stream",
				ContentEncoding: "gzip",
				Path:            "SampleArtifacts/_/X.txt",
			},
		})
}

func TestFileArtifactWithContentEncoding(t *testing.T) {
	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{
			{
				Expires:	inAnHour,
				Path:		"SampleArtifacts/_/X.txt",
				Type:		"file",
				Name:		"public/_/X.txt",
				ContentType:	"text/plain; charset=utf-8",
			},
			{
				Expires:	inAnHour,
				Path:		"SampleArtifacts/b/c/d.jpg",
				Type:		"file",
				Name:		"public/b/c/d.jpg",
				ContentType:	"image/jpeg",
			},
			{
				Expires:		inAnHour,
				Path:			"SampleArtifacts/_/X.txt",
				Type:			"file",
				Name:			"public/_/X.txt",
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"identity",
			},
			{
				Expires:		inAnHour,
				Path:			"SampleArtifacts/b/c/d.jpg",
				Type:			"file",
				Name:			"public/b/c/d.jpg",
				ContentType:		"image/jpeg",
				ContentEncoding:	"identity",
			},
			{
				Expires:		inAnHour,
				Path:			"SampleArtifacts/_/X.txt",
				Type:			"file",
				Name:			"public/_/X.txt",
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"gzip",
			},
			{
				Expires:		inAnHour,
				Path:			"SampleArtifacts/b/c/d.jpg",
				Type:			"file",
				Name:			"public/b/c/d.jpg",
				ContentType:		"image/jpeg",
				ContentEncoding:	"gzip",
			},
		},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/_/X.txt",
					Expires:	inAnHour,
				},
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"gzip",
				Path:			"SampleArtifacts/_/X.txt",
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/b/c/d.jpg",
					Expires:	inAnHour,
				},
				ContentType:		"image/jpeg",
				ContentEncoding:	"identity",
				Path:			"SampleArtifacts/b/c/d.jpg",
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/_/X.txt",
					Expires:	inAnHour,
				},
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"identity",
				Path:			"SampleArtifacts/_/X.txt",
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/b/c/d.jpg",
					Expires:	inAnHour,
				},
				ContentType:		"image/jpeg",
				ContentEncoding:	"identity",
				Path:			"SampleArtifacts/b/c/d.jpg",
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/_/X.txt",
					Expires:	inAnHour,
				},
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"gzip",
				Path:			"SampleArtifacts/_/X.txt",
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/b/c/d.jpg",
					Expires:	inAnHour,
				},
				ContentType:		"image/jpeg",
				ContentEncoding:	"gzip",
				Path:			"SampleArtifacts/b/c/d.jpg",
			},
		})
}

func TestDirectoryArtifactWithNames(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{
			{
				Expires: inAnHour,
				Path:    "SampleArtifacts",
				Type:    "directory",
				Name:    "public/b/c",
			},
		},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/%%%/v/X",
					Expires: inAnHour,
				},
				ContentType:     "application/octet-stream",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "%%%", "v", "X"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/_/X.txt",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "_", "X.txt"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/b/c/d.jpg",
					Expires: inAnHour,
				},
				ContentType:     "image/jpeg",
				ContentEncoding: "identity",
				Path:            filepath.Join("SampleArtifacts", "b", "c", "d.jpg"),
			},
		})
}

func TestDirectoryArtifactWithContentType(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{
			{
				Expires:     inAnHour,
				Path:        "SampleArtifacts",
				Type:        "directory",
				Name:        "public/b/c",
				ContentType: "text/plain; charset=utf-8",
			},
		},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/%%%/v/X",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "%%%", "v", "X"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/_/X.txt",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "_", "X.txt"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/b/c/d.jpg",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "identity",
				Path:            filepath.Join("SampleArtifacts", "b", "c", "d.jpg"),
			},
		})
}

func TestDirectoryArtifactWithContentEncoding(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{
			{
				Expires:		inAnHour,
				Path:			"SampleArtifacts",
				Type:			"directory",
				Name:			"public/b/c",
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"identity",
			},
			{
				Expires:		inAnHour,
				Path:			"SampleArtifacts",
				Type:			"directory",
				Name:			"public/b/c",
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"gzip",
			},
		},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/%%%/v/X",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "identity",
				Path:            filepath.Join("SampleArtifacts", "%%%", "v", "X"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/_/X.txt",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "identity",
				Path:            filepath.Join("SampleArtifacts", "_", "X.txt"),
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/b/c/b/c/d.jpg",
					Expires:	inAnHour,
				},
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"identity",
				Path:			filepath.Join("SampleArtifacts", "b", "c", "d.jpg"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/%%%/v/X",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "%%%", "v", "X"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "public/b/c/_/X.txt",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "_", "X.txt"),
			},
			&S3Artifact{
				BaseArtifact:		&BaseArtifact{
					Name:		"public/b/c/b/c/d.jpg",
					Expires:	inAnHour,
				},
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"gzip",
				Path:			filepath.Join("SampleArtifacts", "b", "c", "d.jpg"),
			},
		})
}

// See the testdata/SampleArtifacts subdirectory of this project. This
// simulates adding it as a directory artifact in a task payload, and checks
// that all files underneath this directory are discovered and created as s3
// artifacts.
func TestDirectoryArtifacts(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{{
			Expires: inAnHour,
			Path:    "SampleArtifacts",
			Type:    "directory",
		}},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "SampleArtifacts/%%%/v/X",
					Expires: inAnHour,
				},
				ContentType:     "application/octet-stream",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "%%%", "v", "X"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "SampleArtifacts/_/X.txt",
					Expires: inAnHour,
				},
				ContentType:     "text/plain; charset=utf-8",
				ContentEncoding: "gzip",
				Path:            filepath.Join("SampleArtifacts", "_", "X.txt"),
			},
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "SampleArtifacts/b/c/d.jpg",
					Expires: inAnHour,
				},
				ContentType:     "image/jpeg",
				ContentEncoding: "identity",
				Path:            filepath.Join("SampleArtifacts", "b", "c", "d.jpg"),
			},
		})
}

// Task payload specifies a file artifact which doesn't exist on worker
func TestMissingFileArtifact(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{{
			Expires: inAnHour,
			Path:    t.Name() + "/no_such_file",
			Type:    "file",
		}},

		// what we expect to discover on file system
		[]TaskArtifact{
			&ErrorArtifact{
				BaseArtifact: &BaseArtifact{
					Name:    t.Name() + "/no_such_file",
					Expires: inAnHour,
				},
				Path:    t.Name() + "/no_such_file",
				Message: "Could not read file '" + filepath.Join(taskContext.TaskDir, t.Name(), "no_such_file") + "'",
				Reason:  "file-missing-on-worker",
			},
		})
}

// Task payload specifies a directory artifact which doesn't exist on worker
func TestMissingDirectoryArtifact(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{{
			Expires: inAnHour,
			Path:    t.Name() + "/no_such_dir",
			Type:    "directory",
		}},

		// what we expect to discover on file system
		[]TaskArtifact{
			&ErrorArtifact{
				BaseArtifact: &BaseArtifact{
					Name:    t.Name() + "/no_such_dir",
					Expires: inAnHour,
				},
				Path:    t.Name() + "/no_such_dir",
				Message: "Could not read directory '" + filepath.Join(taskContext.TaskDir, t.Name(), "no_such_dir") + "'",
				Reason:  "file-missing-on-worker",
			},
		})
}

// Task payload specifies a file artifact which is actually a directory on worker
func TestFileArtifactIsDirectory(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{{
			Expires: inAnHour,
			Path:    "SampleArtifacts/b/c",
			Type:    "file",
		}},

		// what we expect to discover on file system
		[]TaskArtifact{
			&ErrorArtifact{
				BaseArtifact: &BaseArtifact{
					Name:    "SampleArtifacts/b/c",
					Expires: inAnHour,
				},
				Path:    "SampleArtifacts/b/c",
				Message: "File artifact '" + filepath.Join(taskContext.TaskDir, "SampleArtifacts", "b", "c") + "' exists as a directory, not a file, on the worker",
				Reason:  "invalid-resource-on-worker",
			},
		})
}

// TestDefaultArtifactExpiry tests that when providing no artifact expiry, task expiry is used
func TestDefaultArtifactExpiry(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{{
			Path: "SampleArtifacts/b/c/d.jpg",
			Type: "file",
		}},

		// what we expect to discover on file system
		[]TaskArtifact{
			&S3Artifact{
				BaseArtifact: &BaseArtifact{
					Name:    "SampleArtifacts/b/c/d.jpg",
					Expires: inAnHour,
				},
				ContentType:     "image/jpeg",
				ContentEncoding: "identity",
				Path:            "SampleArtifacts/b/c/d.jpg",
			},
		},
	)
}

// Task payload specifies a directory artifact which is a regular file on worker
func TestDirectoryArtifactIsFile(t *testing.T) {

	defer setup(t)()
	validateArtifacts(t,

		// what appears in task payload
		[]Artifact{{
			Expires: inAnHour,
			Path:    "SampleArtifacts/b/c/d.jpg",
			Name:    "SampleArtifacts/b/c/d.jpg",
			Type:    "directory",
		}},

		// what we expect to discover on file system
		[]TaskArtifact{
			&ErrorArtifact{
				BaseArtifact: &BaseArtifact{
					Name:    "SampleArtifacts/b/c/d.jpg",
					Expires: inAnHour,
				},
				Path:    "SampleArtifacts/b/c/d.jpg",
				Message: "Directory artifact '" + filepath.Join(taskContext.TaskDir, "SampleArtifacts", "b", "c", "d.jpg") + "' exists as a file, not a directory, on the worker",
				Reason:  "invalid-resource-on-worker",
			},
		})
}

func TestMissingArtifactFailsTest(t *testing.T) {

	defer setup(t)()

	expires := tcclient.Time(time.Now().Add(time.Minute * 30))

	payload := GenericWorkerPayload{
		Command:    append(helloGoodbye()),
		MaxRunTime: 30,
		Artifacts: []Artifact{
			{
				Path:    "Nonexistent/art i fact.txt",
				Expires: expires,
				Type:    "file",
			},
		},
	}

	td := testTask(t)

	_ = submitAndAssert(t, td, payload, "failed", "failed")
}

func TestInvalidContentEncoding(t *testing.T) {

	defer setup(t)()

	expires := tcclient.Time(time.Now().Add(time.Minute * 30))

	command := helloGoodbye()

	payload := GenericWorkerPayload{
		Command:	command,
		MaxRunTime:	30,
		Artifacts:	[]Artifact{
			{
				Path:			"SampleArtifacts/_/X.txt",
				Expires:		expires,
				Type:			"file",
				Name:			"public/_/X.txt",
				ContentType:		"text/plain; charset=utf-8",
				ContentEncoding:	"jpg",
			},
		},
	}
	td := testTask(t)

	_ = submitAndAssert(t, td, payload, "exception", "malformed-payload")

	// check log mentions contentEncoding invalid
	bytes, err := ioutil.ReadFile(filepath.Join(taskContext.TaskDir, logPath))
	if err != nil {
		t.Fatalf("Error when trying to read log file: %v", err)
	}
	logtext := string(bytes)
	if !strings.Contains(logtext, "[taskcluster:error] - artifacts.0.contentEncoding: artifacts.0.contentEncoding must be one of the following: \"identity\", \"gzip\"") {
		t.Fatalf("Was expecting log file to explain that contentEncoding was invalid, but it doesn't: \n%v", logtext)
	}
}

func TestInvalidContentEncodingBlacklisted(t *testing.T) {

	defer setup(t)()

	expires := tcclient.Time(time.Now().Add(time.Minute * 30))

	command := helloGoodbye()

	payload := GenericWorkerPayload{
		Command:	command,
		MaxRunTime:	30,
		Artifacts:	[]Artifact{
			{
				Path:			"SampleArtifacts/b/c/d.jpg",
				Expires:		expires,
				Type:			"file",
				Name:			"public/b/c/d.jpg",
				ContentType:		"image/jpeg",
				ContentEncoding:	"jpg",
			},
		},
	}
	td := testTask(t)

	_ = submitAndAssert(t, td, payload, "exception", "malformed-payload")

	// check log mentions contentEncoding invalid
	bytes, err := ioutil.ReadFile(filepath.Join(taskContext.TaskDir, logPath))
	if err != nil {
		t.Fatalf("Error when trying to read log file: %v", err)
	}
	logtext := string(bytes)
	if !strings.Contains(logtext, "[taskcluster:error] - artifacts.0.contentEncoding: artifacts.0.contentEncoding must be one of the following: \"identity\", \"gzip\"") {
		t.Fatalf("Was expecting log file to explain that contentEncoding was invalid, but it doesn't: \n%v", logtext)
	}
}

func TestEmptyContentEncoding(t *testing.T){

	defer setup(t)()

	td := testTask(t)
	td.Payload = json.RawMessage(`
{
  "command": [`+ rawHelloGoodbye() +`],
  "maxRunTime": 30,
  "artifacts": [
    {
      "path": "SampleArtifacts/b/c/d.jpg",
      "expires": "` + tcclient.Time(time.Now().Add(time.Minute * 30)).String() + `",
      "type": "file",
      "name": "public/b/c/d.jpg",
      "contentType": "image/jpeg",
      "contentEncoding": ""
    }
  ]
}`)

	_ = submitAndAssert(t, td, GenericWorkerPayload{}, "exception", "malformed-payload")

	// check log mentions contentEncoding invalid
	bytes, err := ioutil.ReadFile(filepath.Join(taskContext.TaskDir, logPath))
	if err != nil {
		t.Fatalf("Error when trying to read log file: %v", err)
	}
	logtext := string(bytes)
	if !strings.Contains(logtext, "[taskcluster:error] - artifacts.0.contentEncoding: artifacts.0.contentEncoding must be one of the following: \"identity\", \"gzip\"") {
		t.Fatalf("Was expecting log file to explain that contentEncoding was invalid, but it doesn't: \n%v", logtext)
	}
}
