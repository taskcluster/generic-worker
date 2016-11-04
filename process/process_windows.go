package process

import (
	"fmt"
	"io"
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
	case r.SuccessCode == 0:
		return SUCCESS
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
	return GetVerdict(r).String()
}

func (c *Command) String() string {
	return *c.Cmd.CommandLine
}

func (c *Command) Execute() (r *Result) {
	result, err := c.Subprocess.Execute()
	return &Result{
		SubprocessResult: result,
		SystemError:      err,
	}
}

func NewCommand(commandLine string, workingDirectory *string, env *[]string, username, password string, timeLimit time.Duration) (*Command, error) {
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
	command := &Command{
		&subprocess.Subprocess{
			TimeQuantum: time.Second / 4,
			Cmd: &subprocess.CommandLine{
				ApplicationName: nil,
				CommandLine:     &commandLine,
				Parameters:      nil,
			},
			CurrentDirectory:    workingDirectory,
			TimeLimit:           0,
			HardTimeLimit:       timeLimit,
			MemoryLimit:         0,
			CheckIdleness:       false,
			RestrictUi:          false,
			ProcessAffinityMask: 0,
			NoJob:               true,
			Environment:         env,
			StdIn:               nil,
			StdOut:              nil,
			StdErr:              nil,
			JoinStdOutErr:       true,
			Options: &subprocess.PlatformOptions{
				Desktop: desktopName,
			},
			Login: loginInfo,
		},
	}
	return command, nil
}

func (c *Command) Kill() error {
	return nil
}

func (c *Command) DirectOutput(writer io.Writer) {
}
