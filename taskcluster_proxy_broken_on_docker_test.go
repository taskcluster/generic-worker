// +build !dockerEngine

package main

import (
	"os"
	"testing"
)

func TestTaskclusterProxy(t *testing.T) {
	if os.Getenv("TASKCLUSTER_PROXY_URL") == "" {
		if os.Getenv("TASKCLUSTER_CLIENT_ID") == "" ||
			os.Getenv("TASKCLUSTER_ACCESS_TOKEN") == "" ||
			os.Getenv("TASKCLUSTER_ROOT_URL") == "" {
			t.Skip("Skipping test since TASKCLUSTER_{CLIENT_ID,ACCESS_TOKEN,ROOT_URL} env vars not set")
		}
	}

	defer setup(t)()
	payload := GenericWorkerPayload{
		Command: append(
			append(
				goEnv(),
				// long enough to reclaim and get new credentials
				sleep(12)...,
			),
			goRun(
				"curlget.go",
				// note that curlget.go supports substituting the proxy URL from its runtime environment
				"TASKCLUSTER_PROXY_URL/queue/v1/task/KTBKfEgxR5GdfIIREQIvFQ/runs/0/artifacts/SampleArtifacts/_/X.txt",
			)...,
		),
		MaxRunTime: 60,
		Env:        map[string]string{},
		Features: FeatureFlags{
			TaskclusterProxy: true,
		},
	}
	for _, envVar := range []string{
		"PATH",
		"GOPATH",
		"GOROOT",
	} {
		if v, exists := os.LookupEnv(envVar); exists {
			payload.Env[envVar] = v
		}
	}
	td := testTask(t)
	td.Scopes = []string{"queue:get-artifact:SampleArtifacts/_/X.txt"}
	reclaimEvery5Seconds = true
	taskID := submitAndAssert(t, td, payload, "completed", "completed")
	reclaimEvery5Seconds = false

	expectedArtifacts := ExpectedArtifacts{
		"public/logs/live_backing.log": {
			Extracts: []string{
				"test artifact",
				"Successfully refreshed taskcluster-proxy credentials",
			},
			ContentType:     "text/plain; charset=utf-8",
			ContentEncoding: "gzip",
			Expires:         td.Expires,
		},
	}

	expectedArtifacts.Validate(t, taskID, 0)
}
