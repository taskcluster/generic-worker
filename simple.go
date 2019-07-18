// +build simple

package main

import (
	"log"

	"github.com/taskcluster/generic-worker/process"
)

const (
	engine = "simple"
)

func secureConfigFile() {
	log.Print("WARNING: can't secure generic-worker config file")
}

func (task *TaskRun) generateCommand(index int) error {
	var err error
	task.Commands[index], err = process.NewCommand(task.Payload.Command[index], taskContext.TaskDir, task.EnvVars())
	if err != nil {
		return err
	}
	task.logMux.RLock()
	defer task.logMux.RUnlock()
	task.Commands[index].DirectOutput(task.logWriter)
	return nil
}
