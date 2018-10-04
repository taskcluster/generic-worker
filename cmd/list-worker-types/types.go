package main

import (
	"encoding/json"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

// *********************************************************
// These type definitions are copied from:
// https://github.com/taskcluster/generic-worker/blob/ec86473df8dba68631a50af98e5af7d44d7e1717/generated_windows.go#L40-L201
// *********************************************************

type (
	// This schema defines the structure of the `payload` property referred to in a
	// Taskcluster Task definition.
	GenericWorkerPayload struct {

		// Artifacts to be published.
		//
		// Since: generic-worker 1.0.0
		Artifacts []struct {

			// Explicitly set the value of the HTTP `Content-Type` response header when the artifact(s)
			// is/are served over HTTP(S). If not provided (this property is optional) the worker will
			// guess the content type of artifacts based on the filename extension of the file storing
			// the artifact content. It does this by looking at the system filename-to-mimetype mappings
			// defined in the Windows registry. Note, setting `contentType` on a directory artifact will
			// apply the same contentType to all files contained in the directory.
			//
			// See [mime.TypeByExtension](https://godoc.org/mime#TypeByExtension).
			//
			// Since: generic-worker 10.4.0
			ContentType string `json:"contentType,omitempty"`

			// Date when artifact should expire must be in the future, no earlier than task deadline, but
			// no later than task expiry. If not set, defaults to task expiry.
			//
			// Since: generic-worker 1.0.0
			Expires tcclient.Time `json:"expires,omitempty"`

			// Name of the artifact, as it will be published. If not set, `path` will be used.
			// Conventionally (although not enforced) path elements are forward slash separated. Example:
			// `public/build/a/house`. Note, no scopes are required to read artifacts beginning `public/`.
			// Artifact names not beginning `public/` are scope-protected (caller requires scopes to
			// download the artifact). See the Queue documentation for more information.
			//
			// Since: generic-worker 8.1.0
			Name string `json:"name,omitempty"`

			// Relative path of the file/directory from the task directory. Note this is not an absolute
			// path as is typically used in docker-worker, since the absolute task directory name is not
			// known when the task is submitted. Example: `dist\regedit.exe`. It doesn't matter if
			// forward slashes or backslashes are used.
			//
			// Since: generic-worker 1.0.0
			Path string `json:"path"`

			// Artifacts can be either an individual `file` or a `directory` containing
			// potentially multiple files with recursively included subdirectories.
			//
			// Since: generic-worker 1.0.0
			//
			// Possible values:
			//   * "file"
			//   * "directory"
			Type string `json:"type"`
		} `json:"artifacts,omitempty"`

		// One entry per command (consider each entry to be interpreted as a full line of
		// a Windows™ .bat file). For example:
		// ```
		// [
		//   "set",
		//   "echo hello world > hello_world.txt",
		//   "set GOPATH=C:\\Go"
		// ]
		// ```
		//
		// Since: generic-worker 0.0.1
		Command []string `json:"command"`

		// Env vars must be string to __string__ mappings (not number or boolean). For example:
		// ```
		// {
		//   "PATH": "C:\\Windows\\system32;C:\\Windows",
		//   "GOOS": "windows",
		//   "FOO_ENABLE": "true",
		//   "BAR_TOTAL": "3"
		// }
		// ```
		//
		// Since: generic-worker 0.0.1
		Env map[string]string `json:"env,omitempty"`

		// Feature flags enable additional functionality.
		//
		// Since: generic-worker 5.3.0
		Features struct {

			// An artifact named `public/chainOfTrust.json.asc` should be generated
			// which will include information for downstream tasks to build a level
			// of trust for the artifacts produced by the task and the environment
			// it ran in.
			//
			// Since: generic-worker 5.3.0
			ChainOfTrust bool `json:"chainOfTrust,omitempty"`

			// The taskcluster proxy provides an easy and safe way to make authenticated
			// taskcluster requests within the scope(s) of a particular task. See
			// [the github project](https://github.com/taskcluster/taskcluster-proxy) for more information.
			//
			// Since: generic-worker 10.6.0
			TaskclusterProxy bool `json:"taskclusterProxy,omitempty"`
		} `json:"features,omitempty"`

		// Maximum time the task container can run in seconds.
		//
		// Since: generic-worker 0.0.1
		//
		// Mininum:    1
		// Maximum:    86400
		MaxRunTime int64 `json:"maxRunTime"`

		// Directories and/or files to be mounted.
		//
		// Since: generic-worker 5.4.0
		Mounts []Mount `json:"mounts,omitempty"`

		// A list of OS Groups that the task user should be a member of. Requires
		// scope `generic-worker:os-group:<os-group>` for each group listed.
		//
		// Since: generic-worker 6.0.0
		OSGroups []string `json:"osGroups,omitempty"`

		// Specifies an artifact name for publishing RDP connection information.
		//
		// Since this is potentially sensitive data, care should be taken to publish
		// to a suitably locked down path, such as
		// `login-identity/<login-identity>/rdpinfo.json` which is only readable for
		// the given login identity (for example
		// `login-identity/mozilla-ldap/pmoore@mozilla.com/rdpInfo.txt`). See the
		// [artifact namespace guide](https://docs.taskcluster.net/manual/design/namespaces#artifacts) for more information.
		//
		// Use of this feature requires scope
		// `generic-worker:allow-rdp:<provisionerId>/<workerType>` which must be
		// declared as a task scope.
		//
		// The RDP connection data is published during task startup so that a user
		// may interact with the running task.
		//
		// The task environment will be retained for 12 hours after the task
		// completes, to enable an interactive user to perform investigative tasks.
		// After these 12 hours, the worker will delete the task's Windows user
		// account, and then continue with other tasks.
		//
		// No guarantees are given about the resolution status of the interactive
		// task, since the task is inherently non-reproducible and no automation
		// should rely on this value.
		//
		// Since: generic-worker 10.5.0
		RdpInfo string `json:"rdpInfo,omitempty"`

		// URL of a service that can indicate tasks superseding this one; the current `taskId`
		// will be appended as a query argument `taskId`. The service should return an object with
		// a `supersedes` key containing a list of `taskId`s, including the supplied `taskId`. The
		// tasks should be ordered such that each task supersedes all tasks appearing later in the
		// list.
		//
		// See [superseding](https://docs.taskcluster.net/reference/platform/taskcluster-queue/docs/superseding) for more detail.
		//
		// Since: generic-worker 10.2.2
		SupersederURL string `json:"supersederUrl,omitempty"`
	}

	Mount json.RawMessage
)
