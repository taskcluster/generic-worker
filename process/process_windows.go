package process

import (
	"io"
	"time"

	"github.com/contester/runlib/platform"
	"github.com/contester/runlib/subprocess"
)

type Command struct {
	*subprocess.Subprocess
}

type Result struct {
	ExitError   error
	SystemError error
}

func (r *Result) Succeeded() bool {
	return true
}

func (r *Result) Failed() bool {
	return false
}

func (r *Result) Crashed() bool {
	return false
}

func (r *Result) String() string {
	return ""
}

func (c *Command) String() string {
	return ""
}

func (c *Command) Execute() (r *Result) {
	return nil
}

func NewCommand(commandLine string, workingDirectory *string, env *[]string, username, password string, timeLimit time.Duration) (*Command, error) {
	desktop, err := platform.CreateContesterDesktopStruct()
	if err != nil {
		return nil, err
	}
	loginInfo, err := subprocess.NewLoginInfo(username, password)
	if err != nil {
		return nil, err
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
				Desktop: desktop.DesktopName,
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
