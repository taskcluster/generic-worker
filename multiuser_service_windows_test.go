package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/windows/svc"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/taskcluster/slugid-go/slugid"
)

func unprivilegedRecovery(t *testing.T) {
	if r := recover(); r != nil {
		t.Log("A panic can occur if tests are run as an unprivileged user.")
		t.Log("Disable tests that require administrator privileges by setting `SKIP_ADMINISTRATOR_TESTS` in your environment")
		t.Fatalf("Caught panic: %v", r)
	}
}

func setupService(t *testing.T, name string, configureEventlog bool) {
	// doesn't have to be real
	path := os.Args[0]
	args := []string{}
	err := configureService(name, path, args)
	if err != nil {
		t.Fatal(err)
	}
	if configureEventlog {
		err = configureEventlogSource(name, path)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func cleanupService(t *testing.T, name string) {
	// remove service
	err := removeService(name)
	if err != nil {
		t.Fatal(err)
	}
}

// requires elevated privileges
func TestConfigureAndRemoveService(t *testing.T) {
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	name := "generic-worker-" + slugid.Nice()
	setupService(t, name, true)

	// service manager
	m, err := mgr.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Disconnect()

	// check for service
	_, err = m.OpenService(name)
	if err != nil {
		t.Fatalf("Did not find expected service %q: %v", name, err)
	}

	// check for eventlog source
	elog, err := eventlog.Open(name)
	if err != nil {
		t.Fatalf("Did not find expected eventlog source %q: %v", name, err)
	}
	elog.Close()

	// usually the service is marked for deletion
	// but not actually removed until reboot

	cleanupService(t, name)

	// TODO these don't work
	// if we dip into the registry to remove the service
	// we can actually verify it right after removal

	// // verify service is removed
	// _, err = m.OpenService(name)
	// if err == nil {
	// 	t.Fatalf("Found service %q after it should have been removed", name)
	// }

	// // verify eventlog is removed
	// _, err = eventlog.Open(name)
	// if err == nil {
	// 	t.Fatalf("Found eventlog source %q after it should have been removed", name)
	// }
}

type brokenWriter struct{}

func (w brokenWriter) Write(bs []byte) (int, error) {
	return -1, fmt.Errorf("broken writer is broken")
}

func TestRunServiceWithBrokenWriter(t *testing.T) {
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	name := "generic-worker-" + slugid.Nice()
	setupService(t, name, true)
	defer cleanupService(t, name)

	// use brokenWriter
	logWriter = brokenWriter{}

	// now we try runService in non-interactive mode
	// which should attempt to use brokenWriter and fail
	// with CANT_LOG_PROPERLY
	exitCode := runService(name, false)
	// as expected, there was an error
	if exitCode != CANT_LOG_PROPERLY {
		t.Fatalf("Expected runService() to exit with CANT_LOG_PROPERLY, got: %v", exitCode)
	}
}

// https://godoc.org/golang.org/x/sys/windows#SERVICE_STOPPED
// SERVICE_STOPPED          = 1
// SERVICE_START_PENDING    = 2
// SERVICE_STOP_PENDING     = 3
// SERVICE_RUNNING          = 4
// SERVICE_CONTINUE_PENDING = 5
// SERVICE_PAUSE_PENDING    = 6
// SERVICE_PAUSED           = 7
// SERVICE_NO_CHANGE        = 0xffffffff
func receiveStateOrTimeout(t *testing.T, c <-chan svc.Status, expected svc.State) {
	select {
	case <-time.After(time.Second * 5):
		t.Fatalf("Timeout waiting for status %#v", expected)
	case state := <-c:
		if state.State != expected {
			t.Fatalf("Expected state %#v, got Status %#v", expected, state)
		}
	}
}

func TestWindowsServiceInteraction(t *testing.T) {
	defer setup(t)()

	// Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32)
	r := make(chan svc.ChangeRequest, 1)
	c := make(chan svc.Status, 1)

	s := windowsService{}

	exitChan := make(chan ExitCode, 1)
	go func() {
		_, e := s.Execute([]string{}, r, c)
		exitChan <- ExitCode(e)
	}()

	receiveStateOrTimeout(t, c, svc.StartPending)
	receiveStateOrTimeout(t, c, svc.Running)

	// send Interrogate
	r <- svc.ChangeRequest{Cmd: svc.Interrogate}
	t.Log("Sent Interrogate ChangeRequest to service")

	// 0 value, not a real State
	receiveStateOrTimeout(t, c, svc.State(0))
	receiveStateOrTimeout(t, c, svc.State(0))

	// send Stop
	r <- svc.ChangeRequest{Cmd: svc.Stop}
	t.Log("Sent Stop ChangeRequest to service")

	receiveStateOrTimeout(t, c, svc.StopPending)
	receiveStateOrTimeout(t, c, svc.Stopped)

	t.Log("Waiting for exit code from Execute()")
	select {
	case <-time.After(time.Second * 60):
		t.Fatalf("Timeout waiting for exit code from Execute()")
	case exitCode := <-exitChan:
		t.Logf("Got exit code %v from Execute()", exitCode)
		if ExitCode(exitCode) != WORKER_STOPPED {
			t.Fatalf("Expected exit code %v, got: %v", WORKER_STOPPED, exitCode)
		}
	}
}
