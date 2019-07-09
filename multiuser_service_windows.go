// +build multiuser

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"gopkg.in/natefinch/lumberjack.v2"
)

// stderr handle is invalid when run as service
var logWriter = io.MultiWriter(ioutil.Discard)

func init() {
	// default for a service is C:\Windows\system32
	dir := path.Dir(os.Args[0])
	err := os.Chdir(dir)
	if err != nil {
		exitOnError(INTERNAL_ERROR, err, "Unable to chdir to %q", dir)
	}
	log.SetOutput(logWriter)
	manageLogFile(filepath.Join(path.Dir(os.Args[0]), "generic-worker.log"))
}

func manageLogFile(logPath string) {
	logWriter = io.MultiWriter(
		logWriter,
		&lumberjack.Logger{
			Filename:   logPath,
			MaxBackups: 10,
			MaxSize:    20,   // megabytes
			MaxAge:     7,    //days
			Compress:   true, // disabled by default
		},
	)
	log.SetOutput(logWriter)
	// lumberjack opens logfile on first write
	// multiwriter will fail if write to any writer fails
	// so we aggressively handle that scenario
	err := log.Output(2, fmt.Sprintf("Opened logfile %q", logPath))
	if err != nil {
		exitOnError(CANT_LOG_PROPERLY, err, "Unable to log to logfile %q with writer: %v", logPath, logWriter)
	}
}

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
func (*windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// Start worker with interruptChan
	interruptChan := make(chan os.Signal, 1)
	exitChan := make(chan uint32, 1)

	go func() {
		exitCode = uint32(RunWorker(interruptChan))
		// kill the service
		interruptChan <- os.Interrupt
		exitChan <- exitCode
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		// we send this when RunWorker exits
		case <-interruptChan:
			break loop
		// received change request
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				log.Printf("Shutting down, received %v", c)
				interruptChan <- os.Interrupt
				break loop
			default:
				log.Printf("Unexpected control request #%d", c)
			}
		}
	}
	changes <- svc.Status{State: svc.Stopped}
	exitCode = <-exitChan
	return true, exitCode
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
		if err != nil || elog == nil {
			writeCrashFile(CANT_LOG_PROPERLY, err, "Could not open eventlog %q", name)
			return CANT_LOG_PROPERLY
		}
	}
	logWriter = io.MultiWriter(logWriter, elogWrapper{elog})
	log.SetOutput(logWriter)
	// multiwriter will fail if write to any writer fails
	// so we aggressively handle that scenario
	err = log.Output(2, fmt.Sprintf("Wrote to eventlog %q successfully", name))
	if err != nil {
		writeCrashFile(CANT_LOG_PROPERLY, err, "Unable to log to eventlog %q with writer: %v", name, logWriter)
		return CANT_LOG_PROPERLY
	}
	defer elog.Close()

	log.Printf("Starting service %q", name)
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &windowsService{})
	if err != nil {
		writeCrashFile(INTERNAL_ERROR, err, "Failed to start service %q", name)
		return CANT_LOG_PROPERLY
	}
	// use Output to use all configured loggers and handle err
	// io.MultiWriter fails for _all_ writers if any fail
	// this helps us catch that, as normal log.Print* calls
	// silently eat errors
	err = log.Output(2, fmt.Sprintf("Stopped service %q", name))
	if err != nil {
		writeCrashFile(CANT_LOG_PROPERLY, err, "Unable to log to one or more log outputs, configured writer: %v", logWriter)
		return CANT_LOG_PROPERLY
	}
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
		return fmt.Errorf("Service %s already exists", name)
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
	err = configureService(name, exePath, args)
	if err != nil {
		return err
	}
	log.Printf("Successfully configured service %q.", name)
	err = configureEventlogSource(name, exePath)
	if err != nil {
		return err
	}
	log.Printf("Successfully configured eventlog source %q.", name)

	// start service manually in order to fail fast
	err = s.Start(args...)
	if err != nil {
		return fmt.Errorf("Error starting service %q: %v", name, err)
	}
	return nil
}

func configureService(name, exePath string, args []string) error {
	config := mgr.Config{
		DisplayName: name,
		Description: "A taskcluster worker that runs on all mainstream platforms",
		// run as LocalSystem because we call WTSQueryUserToken
		ServiceStartName: "LocalSystem",
		ServiceType:      windows.SERVICE_WIN32_OWN_PROCESS | windows.SERVICE_INTERACTIVE_PROCESS,
		StartType:        mgr.StartAutomatic,
	}
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
	return nil
}

func configureEventlogSource(name, exePath string) error {
	dir := filepath.Dir(exePath)

	// minimal eventlog message file
	// https://docs.microsoft.com/en-us/windows/desktop/eventlog/message-files
	messageFileText := `MessageIdTypedef=DWORD
SeverityNames=(Informational=0x1:STATUS_SEVERITY_INFORMATIONAL)
FacilityNames=(System=0x0:FACILITY_SYSTEM)

MessageId=0x1
Severity=Informational
Facility=System
Language=English`

	eventlogMessageFilepath := filepath.Join(dir, "eventlog-message-file.txt")
	err := ioutil.WriteFile(eventlogMessageFilepath, []byte(messageFileText), 0644)
	if err != nil {
		return fmt.Errorf("Could not write eventlog message file: %s", err)
	}

	// configure eventlog source
	err = eventlog.Install(
		name,
		eventlogMessageFilepath,
		true,
		eventlog.Error|eventlog.Warning|eventlog.Info,
	)
	if err != nil {
		return fmt.Errorf("Setting up eventlog source failed: %s", err)
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
		return fmt.Errorf("Removing eventlog source failed: %s", err)
	}
	log.Printf("Successfully removed service %q.", name)
	return nil
}
