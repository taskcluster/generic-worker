package main

import (
	"fmt"
	"io"
	"os"
	"testing"

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
		t.Error(err)
	}
}

func cleanupService(t *testing.T, name string) {
	// remove service
	err := removeService(name)
	if err != nil {
		t.Error(err)
	}
}

// requires elevated privileges
func TestInstallAndRemoveService(t *testing.T) {
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	t.Logf(os.LookupEnv("SKIP_ADMINISTRATOR_TESTS"))

	name := slugid.Nice()
	setupService(t, name)

	// service manager
	m, err := mgr.Connect()
	if err != nil {
		t.Error(err)
	}
	defer m.Disconnect()

	// check for service
	_, err = m.OpenService(name)
	if err != nil {
		t.Errorf("Did not find expected service %q: %v", name, err)
	}

	// check for eventlog source
	elog, err := eventlog.Open(name)
	if err != nil {
		t.Errorf("Did not find expected eventlog source %q: %v", name, err)
	}
	elog.Close()

	cleanupService(t, name)

	// verify service is removed
	_, err = m.OpenService(name)
	if err == nil {
		t.Errorf("Found service %q after it should have been removed: %v", name, err)
	}

	// verify eventlog is removed
	_, err = eventlog.Open(name)
	if err == nil {
		t.Errorf("Found eventlog source %q after it should have been removed: %v", name, err)
	}
}

func TestRunServiceWithoutEventlogSource(t *testing.T) {
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	name := slugid.Nice()
	setupService(t, name)

	// remove eventlog source
	err := eventlog.Remove(name)
	if err != nil {
		t.Error(err)
	}

	// now we try runService in non-interactive mode
	// which should attempt to use eventlog and fail
	// with CANT_LOG_PROPERLY

	exitCode := runService(name, false)

	if exitCode != CANT_LOG_PROPERLY {
		t.Errorf("Expected runService() to exit with CANT_LOG_PROPERLY, got %q", exitCode)
	}

	cleanupService(t, name)
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

	name := slugid.Nice()
	setupService(t, name)

	// add brokenWriter
	logWriter = io.MultiWriter(logWriter, brokenWriter{})

	// now we try runService in non-interactive mode
	// which should attempt to use brokenWriter and fail
	// with CANT_LOG_PROPERLY

	exitCode := runService(name, false)

	if exitCode != CANT_LOG_PROPERLY {
		t.Errorf("Expected runService() to exit with CANT_LOG_PROPERLY, got %q", exitCode)
	}

	cleanupService(t, name)
}
