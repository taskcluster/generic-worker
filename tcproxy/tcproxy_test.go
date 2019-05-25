package tcproxy

import (
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"testing"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

// Note, we don't test TestTcProxy in the CI because we don't have explicit
// credentials (we operate through the taskcluster proxy of the CI task).
// However skipping this unit tests doesn't matter, since the proxy is already
// tested in the integration test TestTaskclusterProxy in the parent package.
// Leaving it here is useful for developers that just want to run the unit
// tests in this package.
func TestTcProxy(t *testing.T) {
	if os.Getenv("TASKCLUSTER_CLIENT_ID") == "" ||
		os.Getenv("TASKCLUSTER_ACCESS_TOKEN") == "" ||
		os.Getenv("TASKCLUSTER_ROOT_URL") == "" {
		t.Skip("Skipping test since TASKCLUSTER_{CLIENT_ID,ACCESS_TOKEN,ROOT_URL} env vars not set")
	}

	var executable string
	switch runtime.GOOS {
	case "windows":
		executable = "taskcluster-proxy.exe"
	default:
		executable = "taskcluster-proxy"
	}
	creds := &tcclient.Credentials{
		ClientID:         os.Getenv("TASKCLUSTER_CLIENT_ID"),
		AccessToken:      os.Getenv("TASKCLUSTER_ACCESS_TOKEN"),
		Certificate:      os.Getenv("TASKCLUSTER_CERTIFICATE"),
		AuthorizedScopes: []string{"queue:get-artifact:SampleArtifacts/_/X.txt"},
	}
	ll, err := New(executable, 34569, tcclient.RootURLFromEnvVars(), creds)
	// Do defer before checking err since err could be a different error and
	// process may have already started up.
	defer func() {
		err := ll.Terminate()
		if err != nil {
			t.Fatalf("Failed to terminate taskcluster-proxy process:\n%s", err)
		}
	}()
	if err != nil {
		t.Fatalf("Could not initiate taskcluster-proxy process:\n%s", err)
	}
	res, err := http.Get("http://localhost:34569/queue/v1/task/KTBKfEgxR5GdfIIREQIvFQ/runs/0/artifacts/SampleArtifacts/_/X.txt")
	if err != nil {
		t.Fatalf("Could not hit url to download artifact using taskcluster-proxy: %v", err)
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Could not read artifact using taskcluster-proxy: %v", err)
	}
	if string(data) != "test artifact\n" {
		t.Fatalf("Got incorrect data: %v", string(data))
	}
}
