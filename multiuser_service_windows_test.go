package main

import (
	"fmt"
	"io"
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
		t.Logf("Caught panic: %v", r)
	}
}

func setupService(t *testing.T, name string) {
	// doesn't have to be real
	path := os.Args[0]
	args := []string{}
	err := installService(name, path, args)
	if err != nil {
		t.Fatal(err)
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
func TestInstallAndRemoveService(t *testing.T) {
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	name := "generic-worker-" + slugid.Nice()
	setupService(t, name)

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

	cleanupService(t, name)

	// verify service is removed
	_, err = m.OpenService(name)
	if err == nil {
		t.Fatalf("Found service %q after it should have been removed", name)
	}

	// verify eventlog is removed
	_, err = eventlog.Open(name)
	if err == nil {
		t.Fatalf("Found eventlog source %q after it should have been removed", name)
	}
}

func TestRunServiceWithoutEventlogSource(t *testing.T) {
	defer setup(t)()
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	name := slugid.Nice()
	setupService(t, name)

	// remove eventlog source
	err := eventlog.Remove(name)
	if err != nil {
		t.Fatal(err)
	}

	// now we try runService in non-interactive mode
	// which should attempt to use eventlog and fail
	// with CANT_LOG_PROPERLY

	exitCode := runService(name, false)

	if exitCode != CANT_LOG_PROPERLY {
		t.Fatalf("Expected runService() to exit with CANT_LOG_PROPERLY, got %q", exitCode)
	}

	cleanupService(t, name)
}

type brokenWriter struct{}

func (w brokenWriter) Write(bs []byte) (int, error) {
	return -1, fmt.Errorf("broken writer is broken")
}

func TestRunServiceWithBrokenWriter(t *testing.T) {
	defer setup(t)()
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	name := slugid.Nice()
	setupService(t, name)

	// add brokenWriter
	logWriter = io.MultiWriter(logWriter, brokenWriter{})

	// now we try runService in non-interactive mode
	// which should attempt to use brokenWriter and fail
	// with CANT_LOG_PROPERLY

	exitCode := runService(name, false)

	if exitCode != CANT_LOG_PROPERLY {
		t.Fatalf("Expected runService() to exit with CANT_LOG_PROPERLY, got %q", exitCode)
	}

	cleanupService(t, name)
}

func TestSendWindowsServiceInteraction(t *testing.T) {
	defer setup(t)()

	// Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32)
	r := make(chan svc.ChangeRequest, 1)
	c := make(chan svc.Status, 1)

	s := windowsService{}

	var exitCode uint32
	go func() {
		_, exitCode = s.Execute([]string{}, r, c)
		t.Logf("Got exit code %v from Execute()", exitCode)
	}()

	var status svc.Status
	select {
	case <-time.After(time.Second * 5):
		t.Fatalf("Timeout waiting for status svc.StartPending")
	case status = <-c:
		if status.State != svc.StartPending {
			t.Fatalf("Expected state svc.StartPending, got status %v", status)
		}
	}

	select {
	case <-time.After(time.Second * 5):
		t.Fatalf("Timeout waiting for status svc.Running")
	case status = <-c:
		if status.State != svc.Running {
			t.Fatalf("Expected state svc.Running, got status %v", status)
		}
	}

	<-time.After(1 * time.Second)

	// send Stop
	r <- svc.ChangeRequest{Cmd: svc.Stop}

	fmt.Printf("Sent Stop change request")

	select {
	case <-time.After(time.Second * 5):
		t.Fatalf("Timeout waiting for status svc.StopPending")
	case status = <-c:
		if status.State != svc.StopPending {
			t.Fatalf("Expected state svc.StopPending, got status %v", status)
		}
	}

	select {
	case <-time.After(time.Second * 5):
		t.Fatalf("Timeout waiting for status svc.Stopped")
	case status = <-c:
		if status.State != svc.Stopped {
			t.Fatalf("Expected state svc.Stopped, got status %v", status)
		}
	}

	if ExitCode(exitCode) != WORKER_STOPPED {
		t.Fatalf("Expected exit code %v, got: %v", WORKER_STOPPED, exitCode)
	}
}
