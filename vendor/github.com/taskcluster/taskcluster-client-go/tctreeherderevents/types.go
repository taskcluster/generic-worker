// This source code file is AUTO-GENERATED by github.com/taskcluster/jsonschema2go

package tctreeherderevents

import (
	"encoding/json"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

type (
	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display
	Display struct {

		// Mininum:    1
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display/properties/chunkCount
		ChunkCount int64 `json:"chunkCount,omitempty"`

		// Mininum:    1
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display/properties/chunkId
		ChunkID int64 `json:"chunkId,omitempty"`

		// Min length: 1
		// Max length: 100
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display/properties/groupName
		GroupName string `json:"groupName,omitempty"`

		// Min length: 1
		// Max length: 25
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display/properties/groupSymbol
		GroupSymbol string `json:"groupSymbol"`

		// Min length: 1
		// Max length: 100
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display/properties/jobName
		JobName string `json:"jobName"`

		// Min length: 0
		// Max length: 25
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display/properties/jobSymbol
		JobSymbol string `json:"jobSymbol"`
	}

	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/errors/items
	Error struct {

		// Min length: 1
		// Max length: 255
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/errors/items/properties/line
		Line string `json:"line"`

		// Mininum:    0
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/errors/items/properties/linenumber
		Linenumber int64 `json:"linenumber"`
	}

	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[1]
	GithubPullRequest struct {

		// Possible values:
		//   * "github.com"
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[1]/properties/kind
		Kind string `json:"kind"`

		// This could be the organization or the individual git username
		// depending on who owns the repo.
		//
		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[1]/properties/owner
		Owner string `json:"owner,omitempty"`

		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[1]/properties/project
		Project string `json:"project"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[1]/properties/pullRequestID
		PullRequestID int64 `json:"pullRequestID,omitempty"`

		// Min length: 40
		// Max length: 40
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[1]/properties/revision
		Revision string `json:"revision"`
	}

	// PREFERRED: An HG job that only has a revision.  This is for all
	// jobs going forward.
	//
	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[0]
	HGPush struct {

		// Possible values:
		//   * "hg.mozilla.org"
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[0]/properties/kind
		Kind string `json:"kind"`

		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[0]/properties/project
		Project string `json:"project"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[0]/properties/pushLogID
		PushLogID int64 `json:"pushLogID,omitempty"`

		// Syntax:     ^[0-9a-f]+$
		// Min length: 40
		// Max length: 40
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin/oneOf[0]/properties/revision
		Revision string `json:"revision"`
	}

	// Definition of a single job that can be added to Treeherder
	// Project is determined by the routing key, so we don't need to specify it here.
	//
	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#
	JobDefinition struct {

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine
		BuildMachine Machine `json:"buildMachine,omitempty"`

		// The name of the build system that initiated this content.  Some examples
		// are "buildbot" and "taskcluster".  But this could be any name.  This
		// value will be used in the routing key for retriggering jobs in the
		// publish-job-action task.
		//
		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 25
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/buildSystem
		BuildSystem string `json:"buildSystem"`

		// The job guids that were coalesced to this job.
		//
		// Array items:
		// Syntax:     ^[\w/+-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/coalesced/items
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/coalesced
		Coalesced []string `json:"coalesced,omitempty"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/display
		Display Display `json:"display"`

		// Extra information that Treeherder reads on a best-effort basis
		//
		// Additional properties allowed
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/extra
		Extra json.RawMessage `json:"extra,omitempty"`

		// True indicates this job has been retried.
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/isRetried
		IsRetried bool `json:"isRetried,omitempty"`

		// Definition of the Job Info for a job.  These are extra data
		// fields that go along with a job that will be displayed in
		// the details panel within Treeherder.
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo
		JobInfo JobInfo `json:"jobInfo,omitempty"`

		// Possible values:
		//   * "build"
		//   * "test"
		//   * "other"
		//
		// Default:    "other"
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobKind
		JobKind string `json:"jobKind"`

		// Labels are a dimension of a platform.  The values here can vary wildly,
		// so most strings are valid for this.  The list of labels that are used
		// is maleable going forward.
		//
		// These were formerly known as "Options" within "Option Collections" but
		// calling labels now so they can be understood to be just strings that
		// denotes a characteristic of the job.
		//
		// Some examples of labels that have been used:
		//   opt    Optimize Compiler GCC optimize flags
		//   debug  Debug flags passed in
		//   pgo    Profile Guided Optimization - Like opt, but runs with profiling, then builds again using that profiling
		//   asan   Address Sanitizer
		//   tsan   Thread Sanitizer Build
		//
		// Array items:
		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/labels/items
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/labels
		Labels []string `json:"labels,omitempty"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs
		Logs []Log `json:"logs,omitempty"`

		// One of:
		//   * HGPush
		//   * GithubPullRequest
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/origin
		Origin json.RawMessage `json:"origin"`

		// Description of who submitted the job: gaia | scheduler name | username | email
		//
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/owner
		Owner string `json:"owner,omitempty"`

		// Examples include:
		// -  'b2g'
		// -  'firefox'
		// -  'taskcluster'
		// -  'xulrunner'
		//
		// Min length: 1
		// Max length: 125
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/productName
		ProductName string `json:"productName,omitempty"`

		// Examples include:
		// - scheduled
		// - scheduler
		// - Self-serve: Rebuilt by foo@example.com
		// - Self-serve: Requested by foo@example.com
		// - The Nightly scheduler named 'b2g_mozilla-inbound periodic' triggered this build
		// - unknown
		//
		// Min length: 1
		// Max length: 125
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/reason
		Reason string `json:"reason,omitempty"`

		// fail: A failure
		// exception: An infrastructure error/exception
		// success: Build/Test executed without error or failure
		// canceled: The job was cancelled by a user
		// unknown: When the job is not yet completed
		// superseded: When a job has been superseded by another job
		//
		// Possible values:
		//   * "success"
		//   * "fail"
		//   * "exception"
		//   * "canceled"
		//   * "superseded"
		//   * "unknown"
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/result
		Result string `json:"result,omitempty"`

		// The infrastructure retry iteration on this job.  The number of times this
		// job has been retried by the infrastructure.
		// If it's the 1st time running, then it should be 0. If this is the first
		// retry, it will be 1, etc.
		//
		// Default:    0
		// Mininum:    0
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/retryId
		RetryID int64 `json:"retryId,omitempty"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine
		RunMachine Machine `json:"runMachine,omitempty"`

		// unscheduled: not yet scheduled
		// pending: not yet started
		// running: currently in progress
		// completed: Job ran through to completion
		//
		// Possible values:
		//   * "unscheduled"
		//   * "pending"
		//   * "running"
		//   * "completed"
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/state
		State string `json:"state"`

		// This could just be what was formerly submitted as a job_guid in the
		// REST API.
		//
		// Syntax:     ^[A-Za-z0-9/+-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/taskId
		TaskID string `json:"taskId"`

		// Mininum:    1
		// Maximum:    3
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/tier
		Tier int64 `json:"tier,omitempty"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/timeCompleted
		TimeCompleted tcclient.Time `json:"timeCompleted,omitempty"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/timeScheduled
		TimeScheduled tcclient.Time `json:"timeScheduled,omitempty"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/timeStarted
		TimeStarted tcclient.Time `json:"timeStarted,omitempty"`

		// Message version
		//
		// Possible values:
		//   * 1
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/version
		Version int64 `json:"version"`
	}

	// Definition of the Job Info for a job.  These are extra data
	// fields that go along with a job that will be displayed in
	// the details panel within Treeherder.
	//
	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo
	JobInfo struct {

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo/properties/links
		Links []Link `json:"links"`

		// Plain text description of the job and its state.  Submitted with
		// the final message about a task.
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo/properties/summary
		Summary string `json:"summary"`
	}

	// List of URLs shown as key/value pairs.  Shown as:
	// "<label>: <linkText>" where linkText will be a link to the url.
	//
	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo/properties/links/items
	Link struct {

		// Min length: 1
		// Max length: 70
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo/properties/links/items/properties/label
		Label string `json:"label"`

		// Min length: 1
		// Max length: 125
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo/properties/links/items/properties/linkText
		LinkText string `json:"linkText"`

		// Max length: 512
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/jobInfo/properties/links/items/properties/url
		URL string `json:"url"`
	}

	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items
	Log struct {

		// If true, indicates that the number of errors in the log was too
		// large and not all of those lines are indicated here.
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/errorsTruncated
		ErrorsTruncated bool `json:"errorsTruncated,omitempty"`

		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/name
		Name string `json:"name"`

		// This object defines what is seen in the Treeherder Log Viewer.
		// These values can be submitted here, or they will be generated
		// by Treeherder's internal log parsing process from the
		// submitted log.  If this value is submitted, Treeherder will
		// consider the log already parsed and skip parsing.
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps
		Steps []Step `json:"steps,omitempty"`

		// Min length: 1
		// Max length: 255
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/url
		URL string `json:"url"`
	}

	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine
	Machine struct {

		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 25
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine/properties/architecture
		Architecture string `json:"architecture"`

		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 50
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine/properties/name
		Name string `json:"name,omitempty"`

		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 25
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine/properties/os
		OS string `json:"os"`

		// Syntax:     ^[\w-]+$
		// Min length: 1
		// Max length: 100
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/definitions/machine/properties/platform
		Platform string `json:"platform"`
	}

	// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items
	Step struct {

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/errors
		Errors []Error `json:"errors,omitempty"`

		// Mininum:    0
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/lineFinished
		LineFinished int64 `json:"lineFinished"`

		// Mininum:    0
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/lineStarted
		LineStarted int64 `json:"lineStarted"`

		// Min length: 1
		// Max length: 255
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/name
		Name string `json:"name"`

		// Possible values:
		//   * "success"
		//   * "fail"
		//   * "exception"
		//   * "canceled"
		//   * "unknown"
		//
		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/result
		Result string `json:"result"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/timeFinished
		TimeFinished tcclient.Time `json:"timeFinished"`

		// See https://taskcluster-staging.net/schemas/treeherder/v1/pulse-job.json#/properties/logs/items/properties/steps/items/properties/timeStarted
		TimeStarted tcclient.Time `json:"timeStarted"`
	}
)
