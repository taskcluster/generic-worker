// +build docker

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/taskcluster/generic-worker/dockerworker"
	"github.com/taskcluster/generic-worker/process"
	"github.com/taskcluster/shell"
)

func (task *TaskRun) EnvVars() []string {
	env := make([]string, 0, len(task.Payload.Env))

	for key, value := range task.Payload.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

func platformFeatures() []Feature {
	return []Feature{}
}

type PlatformData struct {
	Image json.RawMessage
}

type TaskContext struct {
	TaskDir string
}

func (task *TaskRun) NewPlatformData() (pd *PlatformData, err error) {
	return &PlatformData{}, nil
}

func (pd *PlatformData) ReleaseResources() error {
	return nil
}

func immediateShutdown(cause string) {
}

func immediateReboot() {
}

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	taskContext = &TaskContext{
		TaskDir: pwd,
	}
}

func (task *TaskRun) prepareCommand(index int) error {
	return nil
}

func (task *TaskRun) generateCommand(index int) error {
	var err error
	task.Commands[index], err = process.NewCommand(
		dockerworker.New(context.Background(), task.Queue, task.TaskID, task.logWriter),
		task.Payload.Command[index],
		task.PlatformData.Image,
		taskContext.TaskDir,
		task.EnvVars(),
	)
	return err
}

func purgeOldTasks() error {
	return nil
}

func install(arguments map[string]interface{}) (err error) {
	return nil
}

func makeDirReadableForTaskUser(task *TaskRun, dir string) error {
	// No user separation yet
	return nil
}

func makeDirUnreadableForTaskUser(task *TaskRun, dir string) error {
	// No user separation yet
	return nil
}

func RenameCrossDevice(oldpath, newpath string) error {
	return nil
}

func (task *TaskRun) formatCommand(index int) string {
	return shell.Escape(task.Payload.Command[index]...)
}

func prepareTaskUser(username string) bool {
	return false
}

func deleteTaskDir(path string) error {
	return nil
}

func defaultTasksDir() string {
	// assume all user home directories are all in same folder, i.e. the parent
	// folder of the current user's home folder
	return filepath.Dir(os.Getenv("HOME"))
}

// N/A for unix - just a windows thing
func AutoLogonCredentials() (string, string) {
	return "", ""
}

func chooseTaskDirName() string {
	return "task_" + strconv.Itoa(int(time.Now().Unix()))
}

func unsetAutoLogon() {
}

func deleteTaskDirs() {
	//removeTaskDirs(config.TasksDir)
}

func GrantSIDFullControlOfInteractiveWindowsStationAndDesktop(sid string) (err error) {
	return fmt.Errorf(
		"Cannot grant %v full control of interactive windows station and desktop; platform %v does not have such entities",
		sid,
		runtime.GOOS,
	)
}
