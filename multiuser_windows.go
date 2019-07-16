// +build multiuser

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	stdlibruntime "runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/taskcluster/generic-worker/process"
	"github.com/taskcluster/generic-worker/runtime"
	gwruntime "github.com/taskcluster/generic-worker/runtime"
	"github.com/taskcluster/generic-worker/win32"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
)

func (task *TaskRun) formatCommand(index int) string {
	return task.Payload.Command[index]
}

func platformFeatures() []Feature {
	return []Feature{
		&RDPFeature{},
		&RunAsAdministratorFeature{}, // depends on (must appear later in list than) OSGroups feature
		// keep chain of trust as low down as possible, as it checks permissions
		// of signing key file, and a feature could change them, so we want these
		// checks as late as possible
		&ChainOfTrustFeature{},
	}
}

func immediateReboot() {
	log.Println("Immediate reboot being issued...")
	cmd := exec.Command("C:\\Windows\\System32\\shutdown.exe", "/r", "/t", "3", "/c", "generic-worker requested reboot")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func immediateShutdown(cause string) {
	log.Println("Immediate shutdown being issued...")
	log.Println(cause)
	cmd := exec.Command("C:\\Windows\\System32\\shutdown.exe", "/s", "/t", "3", "/c", cause)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func deleteDir(path string) error {
	log.Print("Trying to remove directory '" + path + "' via os.RemoveAll(path) call...")
	err := os.RemoveAll(path)
	if err == nil {
		return nil
	}
	log.Print("WARNING: could not delete directory '" + path + "' with os.RemoveAll(path) method")
	log.Printf("%v", err)
	log.Print("Trying to remove directory '" + path + "' via del command...")
	err = runtime.RunCommands(
		false,
		[]string{
			"cmd", "/c", "del", "/s", "/q", "/f", path,
		},
	)
	if err != nil {
		log.Printf("%#v", err)
	}
	return err
}

func (task *TaskRun) generateCommand(index int) error {
	commandName := fmt.Sprintf("command_%06d", index)
	wrapper := filepath.Join(taskContext.TaskDir, commandName+"_wrapper.bat")
	log.Printf("Creating wrapper script: %v", wrapper)
	command, err := process.NewCommand([]string{wrapper}, taskContext.TaskDir, nil, taskContext.pd)
	if err != nil {
		return err
	}
	task.logMux.RLock()
	defer task.logMux.RUnlock()
	command.DirectOutput(task.logWriter)
	task.Commands[index] = command
	return nil
}

func (task *TaskRun) prepareCommand(index int) *CommandExecutionError {
	// In order that capturing of log files works, create a custom .bat file
	// for the task which redirects output to a log file...
	env := filepath.Join(taskContext.TaskDir, "env.txt")
	dir := filepath.Join(taskContext.TaskDir, "dir.txt")
	commandName := fmt.Sprintf("command_%06d", index)
	wrapper := filepath.Join(taskContext.TaskDir, commandName+"_wrapper.bat")
	script := filepath.Join(taskContext.TaskDir, commandName+".bat")
	contents := ":: This script runs command " + strconv.Itoa(index) + " defined in TaskId " + task.TaskID + "..." + "\r\n"
	contents += "@echo off\r\n"

	// At the end of each command we export all the env vars, and import them
	// at the start of the next command. Otherwise env variable changes would
	// be lost. Similarly, we store the current directory at the end of each
	// command, and cd into it at the beginning of the subsequent command. The
	// very first command takes the env settings from the payload, and the
	// current directory is set to the home directory of the newly created
	// user.

	// If this is first command, take env from task payload, and cd into home
	// directory
	if index == 0 {
		envVars := map[string]string{}
		for k, v := range task.Payload.Env {
			envVars[k] = v
		}
		for envVar, envValue := range envVars {
			// log.Printf("Setting env var: %v=%v", envVar, envValue)
			contents += "set " + envVar + "=" + envValue + "\r\n"
		}
		contents += "set TASK_ID=" + task.TaskID + "\r\n"
		contents += "set TASKCLUSTER_ROOT_URL=" + config.RootURL + "\r\n"
		if config.RunTasksAsCurrentUser {
			contents += "set TASK_USER_CREDENTIALS=" + filepath.Join(cwd, "current-task-user.json") + "\r\n"
		}
		contents += "cd \"" + taskContext.TaskDir + "\"" + "\r\n"

		// Otherwise get the env from the previous command
	} else {
		for _, x := range [2][2]string{{env, "set "}, {dir, "cd "}} {
			file, err := os.Open(x[0])
			if err != nil {
				panic(fmt.Errorf("Could not read from file %v\n%v", x[0], err))
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				contents += x[1] + scanner.Text() + "\r\n"
			}

			if err := scanner.Err(); err != nil {
				panic(err)
			}
		}
	}

	// see http://blogs.msdn.com/b/oldnewthing/archive/2008/09/26/8965755.aspx
	// need to explicitly unset as we rely on it later
	contents += "set errorlevel=\r\n"

	// now make sure output is enabled again

	// now call the actual script that runs the command

	// ******************************
	// old version that WROTE TO A FILE:
	//      contents += "call " + script + " > " + absLogFile + " 2>&1" + "\r\n"
	// ******************************
	contents += "call " + script + " 2>&1" + "\r\n"
	contents += "@echo off" + "\r\n"

	// store exit code
	contents += "set tcexitcode=%errorlevel%\r\n"

	// now store env for next command, unless this is the last command
	if index != len(task.Payload.Command)-1 {
		contents += "set > " + env + "\r\n"
		contents += "cd > " + dir + "\r\n"
	}

	// exit with stored exit code
	contents += "exit /b %tcexitcode%\r\n"

	// now generate the .bat script that runs all of this
	err := ioutil.WriteFile(
		wrapper,
		[]byte(contents),
		0755, // note this is mostly ignored on windows
	)
	if err != nil {
		panic(err)
	}

	// Now make the actual task a .bat script
	fileContents := []byte(strings.Join([]string{
		"@echo on",
		task.Payload.Command[index],
	}, "\r\n"))

	err = ioutil.WriteFile(
		script,
		fileContents,
		0755, // note this is mostly ignored on windows
	)
	if err != nil {
		panic(err)
	}

	// log.Printf("Script %q:", script)
	// log.Print("Contents:")
	// log.Print(string(fileContents))

	// log.Printf("Wrapper script %q:", wrapper)
	// log.Print("Contents:")
	// log.Print(contents)

	return nil
}

// Set an environment variable in each command.  This can be called from a feature's
// NewTaskFeature method to set variables for the task.
func (task *TaskRun) setVariable(variable string, value string) error {
	for i := range task.Commands {
		newEnv := []string{fmt.Sprintf("%s=%s", variable, value)}
		combined, err := win32.MergeEnvLists(&task.Commands[i].Cmd.Env, &newEnv)
		if err != nil {
			return err
		}
		task.Commands[i].Cmd.Env = *combined
	}
	return nil
}

func install(arguments map[string]interface{}) (err error) {
	exePath, err := ExePath()
	if err != nil {
		return err
	}
	configFile := convertNilToEmptyString(arguments["--config"])
	switch {
	case arguments["service"]:
		name := convertNilToEmptyString(arguments["--service-name"])
		configureForAWS := arguments["--configure-for-aws"].(bool)
		configureForGCP := arguments["--configure-for-gcp"].(bool)
		return deployService(configFile, name, exePath, configureForAWS, configureForGCP)
	}
	return fmt.Errorf("Unknown install target - only 'service' is allowed")
}

func remove(arguments map[string]interface{}) error {
	switch {
	case arguments["service"]:
		name := convertNilToEmptyString(arguments["--service-name"])
		return removeService(name)
	}
	return fmt.Errorf("Unknown remove target - only 'service' is allowed")
}

func ExePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err = os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func makeFileOrDirReadWritableForUser(recurse bool, dir string, user *gwruntime.OSUser) ([]byte, error) {
	// see http://ss64.com/nt/icacls.html
	return exec.Command("icacls", dir, "/grant:r", user.Name+":(OI)(CI)F").CombinedOutput()
}

func makeDirUnreadableForUser(dir string, user *gwruntime.OSUser) ([]byte, error) {
	// see http://ss64.com/nt/icacls.html
	return exec.Command("icacls", dir, "/remove:g", user.Name).CombinedOutput()
}

// The windows implementation of os.Rename(...) doesn't allow renaming files
// across drives (i.e. copy and delete semantics) - this alternative
// implementation is identical to the os.Rename(...) implementation, but
// additionally sets the flag windows.MOVEFILE_COPY_ALLOWED in order to cater
// for oldpath and newpath being on different drives. See:
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365240(v=vs.85).aspx
func RenameCrossDevice(oldpath, newpath string) (err error) {
	var to, from *uint16
	from, err = syscall.UTF16PtrFromString(oldpath)
	if err != nil {
		return
	}
	to, err = syscall.UTF16PtrFromString(newpath)
	if err != nil {
		return
	}
	// this will work for files and directories on same drive, and even for
	// files on different drives, but not for directories on different drives
	err = windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_COPY_ALLOWED)

	// if we fail, could be a folder that needs to be moved to a different
	// drive - however, check it really is a folder, since otherwise we could
	// end up infinitely recursing between RenameCrossDevice and
	// RenameFolderCrossDevice, since they both call into each other
	if err != nil {
		var fi os.FileInfo
		fi, err = os.Stat(oldpath)
		if err != nil {
			return
		}
		if fi.IsDir() {
			err = RenameFolderCrossDevice(oldpath, newpath)
		}
	}
	return
}

func RenameFolderCrossDevice(oldpath, newpath string) (err error) {
	// recursively move files
	moveFile := func(path string, info os.FileInfo, inErr error) (outErr error) {
		if inErr != nil {
			return inErr
		}
		var relPath string
		relPath, outErr = filepath.Rel(oldpath, path)
		if outErr != nil {
			return
		}
		targetPath := filepath.Join(newpath, relPath)
		if info.IsDir() {
			outErr = os.Mkdir(targetPath, info.Mode())
		} else {
			outErr = RenameCrossDevice(path, targetPath)
		}
		return
	}
	err = filepath.Walk(oldpath, moveFile)
	if err != nil {
		return
	}
	err = os.RemoveAll(oldpath)
	return
}

func (task *TaskRun) addUserToGroups(groups []string) (updatedGroups []string, notUpdatedGroups []string) {
	if len(groups) == 0 {
		return []string{}, []string{}
	}
	for _, group := range groups {
		err := runtime.RunCommands(false, []string{"net", "localgroup", group, "/add", taskContext.User.Name})
		if err == nil {
			updatedGroups = append(updatedGroups, group)
		} else {
			notUpdatedGroups = append(notUpdatedGroups, group)
		}
	}
	return
}

func (task *TaskRun) removeUserFromGroups(groups []string) (updatedGroups []string, notUpdatedGroups []string) {
	if len(groups) == 0 {
		return []string{}, []string{}
	}
	for _, group := range groups {
		err := runtime.RunCommands(false, []string{"net", "localgroup", group, "/delete", taskContext.User.Name})
		if err == nil {
			updatedGroups = append(updatedGroups, group)
		} else {
			notUpdatedGroups = append(notUpdatedGroups, group)
		}
	}
	return
}

func RedirectAppData(hUser syscall.Token, folder string) error {
	RoamingAppDataFolder := filepath.Join(folder, "Roaming")
	LocalAppDataFolder := filepath.Join(folder, "Local")
	err := win32.SetAndCreateFolder(hUser, &win32.FOLDERID_RoamingAppData, RoamingAppDataFolder)
	if err != nil {
		log.Printf("%v", err)
		log.Printf("WARNING: Not able to redirect Roaming App Data folder to %v - IGNORING!", RoamingAppDataFolder)
	}
	err = win32.SetAndCreateFolder(hUser, &win32.FOLDERID_LocalAppData, LocalAppDataFolder)
	if err != nil {
		log.Printf("%v", err)
		log.Printf("WARNING: Not able to redirect Local App Data folder to %v - IGNORING!", LocalAppDataFolder)
	}
	return nil
}

func defaultTasksDir() string {
	return win32.ProfilesDirectory()
}

func UACEnabled() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	var enableLUA uint64
	enableLUA, _, err = k.GetIntegerValue("EnableLUA")
	if err != nil {
		return false
	}
	return enableLUA == 1
}

func rebootBetweenTasks() bool {
	return true
}

func platformTargets(arguments map[string]interface{}) ExitCode {
	switch {
	case arguments["install"]:
		// platform specific...
		err := install(arguments)
		if err != nil {
			log.Printf("failed to install service: %v", err)
			return CANT_INSTALL_GENERIC_WORKER
		}
	case arguments["remove"]:
		err := remove(arguments)
		if err != nil {
			log.Printf("failed to remove service: %v", err)
			return CANT_REMOVE_GENERIC_WORKER
		}
	case arguments["run-service"]:
		dir := convertNilToEmptyString(arguments["--working-directory"])
		// default to generic-worker executable parent dir
		if dir == "" {
			dir = filepath.Dir(os.Args[0])
		}
		err := os.Chdir(dir)
		if err != nil {
			log.Printf("Unable to chdir to %q: %v", dir, err)
			return INTERNAL_ERROR
		}
		handleConfig(arguments)
		name := convertNilToEmptyString(arguments["--service-name"])
		isIntSess, err := svc.IsAnInteractiveSession()
		if err != nil {
			log.Printf("failed to determine if we are running in an interactive session: %v", err)
			return INTERNAL_ERROR
		}
		// debug if interactive session
		return runService(name, isIntSess)
	case arguments["grant-winsta-access"]:
		sid := arguments["--sid"].(string)
		err := GrantSIDFullControlOfInteractiveWindowsStationAndDesktop(sid)
		if err != nil {
			log.Printf("Error granting %v full control of interactive windows station and desktop:", sid)
			log.Printf("%v", err)
			return CANT_GRANT_CONTROL_OF_WINSTA_AND_DESKTOP
		}
	default:
		log.Print("Internal error - no target found to run, yet command line parsing successful")
		return INTERNAL_ERROR
	}
	return 0
}

func SetAutoLogin(user *runtime.OSUser) error {
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`, registry.WRITE)
	if err != nil {
		return fmt.Errorf(`Was not able to create registry key 'SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon' due to %s`, err)
	}
	defer k.Close()
	err = k.SetDWordValue("AutoAdminLogon", 1)
	if err != nil {
		return fmt.Errorf(`Was not able to set registry entry 'SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\AutoAdminLogon' to 1 due to %s`, err)
	}
	err = k.SetStringValue("DefaultUserName", user.Name)
	if err != nil {
		return fmt.Errorf(`Was not able to set registry entry 'SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\DefaultUserName' to %q due to %s`, user.Name, err)
	}
	err = k.SetStringValue("DefaultPassword", user.Password)
	if err != nil {
		return fmt.Errorf(`Was not able to set registry entry 'SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\DefaultPassword' to %q due to %s`, user.Password, err)
	}
	return nil
}

func GrantSIDFullControlOfInteractiveWindowsStationAndDesktop(sid string) (err error) {

	stdlibruntime.LockOSThread()
	defer stdlibruntime.UnlockOSThread()

	var winsta win32.Hwinsta
	if winsta, err = win32.GetProcessWindowStation(); err != nil {
		return
	}

	var winstaName string
	winstaName, err = win32.GetUserObjectName(syscall.Handle(winsta))
	if err != nil {
		return
	}

	var desktop win32.Hdesk
	desktop, err = win32.GetThreadDesktop(win32.GetCurrentThreadId())
	if err != nil {
		return
	}

	var desktopName string
	desktopName, err = win32.GetUserObjectName(syscall.Handle(desktop))
	if err != nil {
		return
	}

	fmt.Printf("Windows Station:   %v\n", winstaName)
	fmt.Printf("Desktop:           %v\n", desktopName)

	var everyone *syscall.SID
	everyone, err = syscall.StringToSid(sid)
	if err != nil {
		return
	}

	err = win32.AddAceToWindowStation(winsta, everyone)
	if err != nil {
		return
	}

	err = win32.AddAceToDesktop(desktop, everyone)
	if err != nil {
		return
	}
	return
}

func PreRebootSetup(nextTaskUser *runtime.OSUser) {
	// set APPDATA
	var loginInfo *process.LoginInfo
	var err error
	loginInfo, err = process.NewLoginInfo(nextTaskUser.Name, nextTaskUser.Password)
	if err != nil {
		panic(err)
	}
	err = RedirectAppData(loginInfo.AccessToken(), filepath.Join(config.TasksDir, nextTaskUser.Name, "AppData"))
	if err != nil {
		panic(err)
	}
	err = loginInfo.Release()
	if err != nil {
		panic(err)
	}
}

func MkdirAllTaskUser(dir string, perms os.FileMode) (err error) {
	return os.MkdirAll(dir, perms)
}
