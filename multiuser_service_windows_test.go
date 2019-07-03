package main

import (
	"os"
	"testing"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/taskcluster/slugid-go/slugid"
)

// requires elevated privileges
func TestInstallAndRemoveService(t *testing.T) {
	if !runningAsAdmin() {
		t.Skipf("Not running as admin, skipping %q", t.Name())
	}

	defer func() {
		if r := recover(); r != nil {
			t.Log("A panic can occur if run as an unprivileged user.")
			t.Log("Caught panic:")
			t.Logf("%v", r)
		}
	}()
	name := slugid.Nice()
	// doesn't have to be real
	path := os.Args[0]
	args := []string{}
	err := installService(name, path, args)
	if err != nil {
		t.Error(err)
	}

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

	// remove service
	err = removeService(name)
	if err != nil {
		t.Error(err)
	}

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
