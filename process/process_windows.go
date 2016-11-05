package process

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/contester/runlib/platform"
	"github.com/contester/runlib/subprocess"
)

type Verdict int

const (
	SUCCESS               = Verdict(0)
	FAIL                  = Verdict(1)
	CRASH                 = Verdict(2)
	TIME_LIMIT_EXCEEDED   = Verdict(3)
	MEMORY_LIMIT_EXCEEDED = Verdict(4)
	IDLE                  = Verdict(5)
	SECURITY_VIOLATION    = Verdict(6)
)

func (v Verdict) String() string {
	switch v {
	case SUCCESS:
		return "SUCCEEDED"
	case FAIL:
		return "FAILED"
	case CRASH:
		return "CRASHED"
	case TIME_LIMIT_EXCEEDED:
		return "TIME_LIMIT_EXCEEDED"
	case MEMORY_LIMIT_EXCEEDED:
		return "MEMORY_LIMIT_EXCEEDED"
	case IDLE:
		return "IDLENESS_LIMIT_EXCEEDED"
	case SECURITY_VIOLATION:
		return "SECURITY_VIOLATION"
	}
	return "FAILED"
}

func GetVerdict(r *Result) Verdict {
	switch {
	case r.SuccessCode == 0 && r.ExitCode == 0:
		return SUCCESS
	case r.SuccessCode == 0 && r.ExitCode != 0:
		return FAIL
	case r.SuccessCode&(subprocess.EF_PROCESS_LIMIT_HIT|subprocess.EF_PROCESS_LIMIT_HIT_POST) != 0:
		return SECURITY_VIOLATION
	case r.SuccessCode&(subprocess.EF_INACTIVE|subprocess.EF_TIME_LIMIT_HARD) != 0:
		return IDLE
	case r.SuccessCode&(subprocess.EF_TIME_LIMIT_HIT|subprocess.EF_TIME_LIMIT_HIT_POST) != 0:
		return TIME_LIMIT_EXCEEDED
	case r.SuccessCode&(subprocess.EF_MEMORY_LIMIT_HIT|subprocess.EF_MEMORY_LIMIT_HIT_POST) != 0:
		return MEMORY_LIMIT_EXCEEDED
	default:
		return CRASH
	}
}

type Command struct {
	*subprocess.Subprocess
	Deadline time.Time
}

type Result struct {
	*subprocess.SubprocessResult
	SystemError error
}

func (r *Result) Succeeded() bool {
	return GetVerdict(r) == SUCCESS
}

func (r *Result) Failed() bool {
	return r.SystemError == nil && GetVerdict(r) != SUCCESS
}

func (r *Result) FailureCause() error {
	return fmt.Errorf("%v\n\nExit code: %v", r.Error, r.ExitCode)
}

func (r *Result) Crashed() bool {
	return r.SystemError != nil
}

func (r *Result) CrashCause() error {
	return r.SystemError
}

func (r *Result) String() string {
	return fmt.Sprintf(`Exit Code: %v
Success Code: %v
%v`, r.ExitCode, r.SuccessCode, GetVerdict(r))
}

func (c *Command) String() string {
	return *c.Cmd.CommandLine
}

func (c *Command) Execute() (r *Result) {
	if !c.Deadline.IsZero() {
		c.HardTimeLimit = c.Deadline.Sub(time.Now())
		if c.HardTimeLimit < 0 {
			log.Printf("WARNING: Deadline %v exceeded before command %v has been executed!", c.Deadline, c)
		}
	}
	result, err := c.Subprocess.Execute()
	return &Result{
		SubprocessResult: result,
		SystemError:      err,
	}
}

func NewCommand(commandLine string, workingDirectory *string, env *[]string, username, password string, deadline time.Time) (*Command, error) {
	var loginInfo *subprocess.LoginInfo
	var desktopName string
	if username != "" {
		desktop, err := platform.CreateContesterDesktopStruct()
		if err != nil {
			return nil, err
		}
		desktopName = desktop.DesktopName
		loginInfo, err = subprocess.NewLoginInfo(username, password)
		if err != nil {
			return nil, err
		}
	}
	maxDuration := deadline.Sub(time.Now())
	command := &Command{
		Subprocess: &subprocess.Subprocess{
			TimeQuantum: time.Second / 4,
			Cmd: &subprocess.CommandLine{
				ApplicationName: nil,
				CommandLine:     &commandLine,
				Parameters:      nil,
			},
			CurrentDirectory:    workingDirectory,
			TimeLimit:           0,
			HardTimeLimit:       maxDuration,
			MemoryLimit:         0,
			CheckIdleness:       false,
			RestrictUi:          false,
			ProcessAffinityMask: 0,
			NoJob:               true,
			Environment:         env,
			StdIn: &subprocess.Redirect{
				Mode: subprocess.REDIRECT_NONE,
			},
			StdOut:        nil,
			StdErr:        nil,
			JoinStdOutErr: true,
			Options: &subprocess.PlatformOptions{
				Desktop: desktopName,
			},
			Login: loginInfo,
		},
		Deadline: deadline,
	}
	return command, nil
}

// For now, I don't see a simple way to terminate the process outside of the
// subprocess library.  However, we can set a time limit, so the only thing we
// can't do is kill a process in response to cancelling of a task. That wasn't
// implemented before, so we haven't lost anything, over the old
// implementation. However, at some point, we should find a way to kill the
// process for when we want to cancel tasks.
func (c *Command) Kill() error {
	return nil
}

func (c *Command) DirectOutput(writer io.Writer) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	c.StdOut = &subprocess.Redirect{
		Mode: subprocess.REDIRECT_PIPE,
		Pipe: w,
	}
	go func() {
		io.Copy(writer, r)
	}()
}
