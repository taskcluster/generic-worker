package process

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/taskcluster/generic-worker/runtime"
	"github.com/taskcluster/runlib/subprocess"
	"github.com/taskcluster/runlib/win32"
)

type Command struct {
	mutex sync.RWMutex
	*exec.Cmd
}

type Result struct {
	SystemError error
	ExitError   *exec.ExitError
	Duration    time.Duration
	Aborted     bool
	KernelTime  time.Duration
	UserTime    time.Duration
}

func (r *Result) Succeeded() bool {
	return r.SystemError == nil && r.ExitError == nil
}

func (r *Result) Failed() bool {
	return (r.SystemError == nil && r.ExitError != nil) || r.Aborted
}

func (r *Result) CrashCause() error {
	return r.SystemError
}

func (r *Result) FailureCause() *exec.ExitError {
	return r.ExitError
}

func (r *Result) Crashed() bool {
	return r.SystemError != nil && !r.Aborted
}

func NewCommand(commandLine []string, workingDirectory string, env []string, loginInfo *subprocess.LoginInfo) (*Command, error) {
	if loginInfo != nil && loginInfo.HUser != 0 {
		environment, err := win32.CreateEnvironment(&env, loginInfo.HUser)
		if err != nil {
			return nil, err
		}
		env = *environment
	}
	cmd := exec.Command(commandLine[0], commandLine[1:]...)
	cmd.Env = env
	cmd.Dir = workingDirectory
	isWindows8OrGreater := win32.IsWindows8OrGreater()
	creationFlags := uint32(win32.CREATE_NEW_PROCESS_GROUP | win32.CREATE_NEW_CONSOLE)
	if !isWindows8OrGreater {
		creationFlags |= win32.CREATE_BREAKAWAY_FROM_JOB
	}
	if loginInfo != nil && loginInfo.HUser != 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Token:         syscall.Token(loginInfo.HUser),
			CreationFlags: creationFlags,
		}
	}
	return &Command{
		Cmd: cmd,
	}, nil
}

// Returns the exit code, or
//  -1 if the process has not exited
//  -2 if the process crashed
//  -3 it could not be established what happened
//  -4 if process was aborted
func (r *Result) ExitCode() int {
	if r.Aborted {
		return -4
	}
	if r.SystemError != nil {
		return -2
	}
	if r.ExitError == nil {
		return 0
	}
	if status, ok := r.ExitError.Sys().(syscall.WaitStatus); ok {
		return status.ExitStatus() // -1 if not exited
	}
	return -3
}

func (c *Command) Execute() (r *Result) {
	r = &Result{}
	started := time.Now()
	c.mutex.Lock()
	err := c.Start()
	c.mutex.Unlock()
	if err != nil {
		r.SystemError = err
		return
	}
	// See https://bugzilla.mozilla.org/show_bug.cgi?id=1462293
	//
	// When a command is killed, only the parent process will be killed, and
	// subprocesses may continue running.
	//
	// If any subprocess inherited console handles from the parent process, the
	// kernel will take care to ensure that those file handles will remain open
	// until no processes are still using them.
	//
	// In the case that the command has io.Reader/io.Writer interfaces attached
	// to any of the console handles (e.g. a io.multiWriter, as used by livelog
	// feature) then internally the go standard library will launch go routines
	// to copy data from the console handles into the io interfaces, when the
	// process is created.  Those go routines will only complete once EOF has
	// been reached, which will only happen when all subprocesses that
	// inherited console handles have completed.
	//
	// c.Wait() waits for those internal go routines to complete, which means
	// that even if the command is aborted, calling c.Wait() may never return,
	// since if the process that was killed had launched a subprocess which
	// continues to run (e.g.  waiting on user input), EOF will never be
	// reached on the file handles, and the go routines will never terminate,
	// and c.Wait() will never return.
	//
	// Therefore we call c.Process.Wait() here rather than c.Wait(), which is
	// guaranteed to return, since it does not wait for the i/o copying to
	// complete which is handled by the internal go routines.
	//
	// IMPORTANT: In the case the command is not aborted, the caller should
	// call c.Wait() after calling c.Execute(), in order to ensure that
	// stderr/stdout buffers are flushed to logs. This has to be the
	// responsibility of the caller, since only it knows if the command was
	// aborted or not.
	state, err := c.Process.Wait()
	finished := time.Now()
	// Round(0) forces wall time calculation instead of monotonic time in case machine slept etc
	r.Duration = finished.Round(0).Sub(started)
	r.UserTime = state.UserTime()
	r.KernelTime = state.SystemTime()
	if err != nil {
		r.SystemError = err
	}
	if !state.Success() {
		r.ExitError = &exec.ExitError{ProcessState: state}
	}
	return
}

func (c *Command) String() string {
	return fmt.Sprintf("%q", c.Args)
}

func (r *Result) String() string {
	return fmt.Sprintf(""+
		"   Exit Code: %v\n"+
		"   User Time: %v\n"+
		" Kernel Time: %v\n"+
		"   Wall Time: %v\n"+
		"      Result: %v",
		r.ExitCode(),
		r.UserTime,
		r.KernelTime,
		r.Duration,
		r.Verdict(),
	)
}

func (r *Result) Verdict() string {
	switch {
	case r.Aborted:
		return "ABORTED"
	case r.ExitError == nil:
		return "SUCCEEDED"
	default:
		return "FAILED"
	}
}

func (c *Command) DirectOutput(writer io.Writer) {
	c.Stdout = writer
	c.Stderr = writer
}

func (c *Command) Kill() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.Process == nil {
		// If process hasn't been started yet, nothing to kill
		return nil
	}
	// tasklist()
	log.Printf("Killing process with ID %v... (%p)", c.Process.Pid, c)
	// defer tasklist()
	defer log.Printf("Process with ID %v killed.", c.Process.Pid)
	return c.Process.Kill()
}

func tasklist() {
	cmd := exec.Command("tasklist")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", stdoutStderr)
}

type LogonSession struct {
	User      *runtime.OSUser
	LoginInfo *subprocess.LoginInfo
}
