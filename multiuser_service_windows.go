// +build multiuser

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// elogWrapper is used to allow eventlog
// to be written to by go's log package
// it eats severity in the process
type elogWrapper struct {
	debug.Log
}

func (e elogWrapper) Write(p []byte) (n int, err error) {
	return len(p), e.Info(1, string(p))
}

type windowsService struct{}

// implements Execute for svc.Handler
func (*windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}

	// Start worker with interruptChan
	interruptChan := make(chan os.Signal, 1)

	go RunWorker(interruptChan)

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Printf("Shutting down, received %v", c)
				interruptChan <- os.Interrupt
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				log.Printf("Unexpected control request #%d", c)
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return false, 0
}

func runService(name string, isDebug bool) ExitCode {
	var err error
	if name == "" {
		name = "Generic Worker"
	}

	var elog debug.Log
	if isDebug {
		log.Printf("Debug mode enabled, not using eventlog.")
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			fmt.Printf("Could not open eventlog: %v", err)
			return INTERNAL_ERROR
		}
		log.SetOutput(elogWrapper{elog})
	}
	defer elog.Close()

	log.Printf("Starting service %q", name)
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &windowsService{})
	if err != nil {
		log.Printf("Service %q failed: %v", name, err))
		return INTERNAL_ERROR
	}
	log.Printf("Stopped service %q", name)
	return 0
}

// deploys the generic worker as a windows service named name
// running as the user LocalSystem
// if the service already exists we skip.
func deployService(configFile, name, exePath string, configureForAWS bool, configureForGCP bool) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}
	args := []string{
		"run-service",
	}
	if configFile != "" {
		args = append(args, "--config", configFile)
	}
	if configureForAWS {
		args = append(args, "--configure-for-aws")
	}
	if configureForGCP {
		args = append(args, "--configure-for-gcp")
	}
	err = installService(name, exePath, args)
	if err != nil {
		return err
	}
	log.Printf("Successfully installed service %q.", name)
	return nil
}

func installService(name, exePath string, args []string) error {
	config := mgr.Config{
		DisplayName: name,
		Description: "A taskcluster worker that runs on all mainstream platforms",
		// run as LocalSystem because we call WTSQueryUserToken
		ServiceStartName: "LocalSystem",
		ServiceType:      windows.SERVICE_WIN32_OWN_PROCESS | windows.SERVICE_INTERACTIVE_PROCESS,
		StartType:        mgr.StartAutomatic,
	}
	dir := filepath.Dir(exePath)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.CreateService(
		name,
		exePath,
		config,
		args...,
	)
	if err != nil {
		return err
	}
	defer s.Close()
	log.Printf("Created service %q with exePath %q, and args %v",
		name, exePath, args)

	// TODO configure an eventlog message file
	// https://docs.microsoft.com/en-us/windows/desktop/eventlog/message-files

	// configure eventlog source
	err = eventlog.Install(
		name,
		filepath.Join(dir, "eventlog-message-file.txt"),
		false,
		eventlog.Error|eventlog.Warning|eventlog.Info,
	)
	if err != nil {
		s.Delete()
		return fmt.Errorf("Setting up eventlog source failed: %s", err)
	}

	// start service manually in order to fail fast
	err = s.Start(args...)
	if err != nil {
		return fmt.Errorf("Error starting service %q: %v", name, err)
	}
	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	log.Printf("Successfully removed service %q.", name)
	return nil
}
