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
		t.Logf("Caught panic: %v", r)
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
	defer cleanupService(t, name)

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

	// this is unrealistic, usually the service is marked for deletion
	// but not actually removed until reboot

	// // hopefully service gets removed
	// <-time.After(2 * time.Second)

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

// func wrapTest(t *testing.T, name string, errorFunc func(*exec.ExitError)) {
// if we are running the test
// as opposed to wrapping the test
// if os.Getenv(t.Name()) == "1" {
// 	name := "generic-worker-" + slugid.Nice()
// 	setupService(t, name, true)
// 	defer cleanupService(t, name)
// 	// use brokenWriter
// 	logWriter = brokenWriter{}
// 	// now we try runService in non-interactive mode
// 	// which should attempt to use brokenWriter and fail
// 	// with CANT_LOG_PROPERLY
// 	runService(name, false)
// 	return
// }

// // wrapping the test

// // exitOnError will be called, so os.Exit() will be called
// // so we need to run this test in a wrapper

// cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
// cmd.Env = append(os.Environ(), fmt.Sprintf("%s=1", t.Name()))
// stderrPipe, err := cmd.StderrPipe()
// if err != nil {
// 	t.Fatal(err)
// }
// stdoutPipe, err := cmd.StdoutPipe()
// if err != nil {
// 	t.Fatalf("Error getting StdoutPipe: %v", err)
// }
// if err := cmd.Start(); err != nil {
// 	t.Fatal(err)
// }
// stderr, err := ioutil.ReadAll(stderrPipe)
// if err != nil {
// 	t.Fatal(err)
// }
// stdout, err := ioutil.ReadAll(stdoutPipe)
// if err != nil {
// 	t.Fatal(err)
// }
// err = cmd.Wait()
// // if nonzero exit code, castable to ExitError
// if exitError, ok := err.(*exec.ExitError); ok && err != nil {
// 	// as expected, there was an error
// 	exitCode := exitError.ExitCode()
// 	if exitCode != int(CANT_LOG_PROPERLY) {
// 		t.Logf("Expected runService() to exit with CANT_LOG_PROPERLY, got: %s", exitError.Error())
// 		t.Logf("stderr: %s", stderr)
// 		t.Logf("stdout: %s", stdout)
// 		t.Fail()
// 	}
// } else {
// 	t.Fatalf("Expected an error from %q: %v", strings.Join(cmd.Args, " "), err)
// }
// }

func TestRunServiceWithBrokenWriter(t *testing.T) {
	defer unprivilegedRecovery(t)

	if !shouldRunAdminTests() {
		t.Skipf("SKIP_ADMINISTRATOR_TESTS set, skipping %q", t.Name())
	}

	// if we are running the test
	// as opposed to wrapping the test
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

func TestWindowsServiceInteraction(t *testing.T) {
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
