<img align="right" src="https://avatars3.githubusercontent.com/u/6257436?s=256" /> A generic worker for [taskcluster](https://tools.taskcluster.net/), written in go.

[![Taskcluster CI Status](https://github.taskcluster.net/v1/repository/taskcluster/generic-worker/master/badge.svg)](https://github.taskcluster.net/v1/repository/taskcluster/generic-worker/master/latest)
[![Linux Build Status](https://img.shields.io/travis/taskcluster/generic-worker.svg?style=flat-square&label=linux+build)](https://travis-ci.org/taskcluster/generic-worker)
[![GoDoc](https://godoc.org/github.com/taskcluster/generic-worker?status.svg)](https://godoc.org/github.com/taskcluster/generic-worker)
[![Coverage Status](https://coveralls.io/repos/taskcluster/generic-worker/badge.svg?branch=master&service=github)](https://coveralls.io/github/taskcluster/generic-worker?branch=master)
[![License](https://img.shields.io/badge/license-MPL%202.0-orange.svg)](http://mozilla.org/MPL/2.0)

# Table of Contents

   * [Introdution](#introdution)
      * [Imperative task payloads](#imperative-task-payloads)
   * [Sandboxing](#sandboxing)
      * [Windows](#windows)
         * [Task user lifecycle](#task-user-lifecycle)
      * [Linux](#linux)
      * [macOS](#macos)
   * [Operating System integration](#operating-system-integration)
      * [Windows](#windows-1)
         * [Creating a task user](#creating-a-task-user)
         * [Setting Known Folder locations](#setting-known-folder-locations)
         * [Configuring auto-logon of task user](#configuring-auto-logon-of-task-user)
         * [Rebooting](#rebooting)
         * [Executing task commands](#executing-task-commands)
   * [Payload format](#payload-format)
   * [Redeployability](#redeployability)
   * [Integrating with AWS / GCE](#integrating-with-aws--gce)
   * [Config bootstrapping](#config-bootstrapping)
   * [Bring your own worker](#bring-your-own-worker)
      * [Windows](#windows-2)
         * [Installing](#installing)
      * [Mac](#mac)
         * [Installing](#installing-1)
      * [Linux - Docker](#linux---docker)
      * [Linux - Native](#linux---native)
         * [Installing](#installing-2)
   * [Administrative tools](#administrative-tools)
      * [Displaying workers](#displaying-workers)
   * [Worker Type Host Definitions](#worker-type-host-definitions)
      * [Updating existing definitions](#updating-existing-definitions)
      * [Modifying definitions](#modifying-definitions)
      * [Creating your own AWS workers outside of this repo](#creating-your-own-aws-workers-outside-of-this-repo)
      * [Puppet](#puppet)
   * [Developing Generic Worker](#developing-generic-worker)
      * [Fetching source](#fetching-source)
      * [Credentials](#credentials)
      * [Running unit tests](#running-unit-tests)
      * [Writing unit tests](#writing-unit-tests)
      * [Including bug numbers in comments](#including-bug-numbers-in-comments)
   * [Releasing Generic Worker](#releasing-generic-worker)
      * [Release script](#release-script)
      * [Publishing schemas](#publishing-schemas)
      * [Testing in Staging](#testing-in-staging)
      * [Rolling out to Production](#rolling-out-to-production)
      * [Writing release notes (README.md, release page, ...)](#writing-release-notes-readmemd-release-page-)
   * [Repository layout](#repository-layout)
   * [Downloading generic-worker binary release](#downloading-generic-worker-binary-release)
   * [Building generic-worker from source](#building-generic-worker-from-source)
   * [Acquire taskcluster credentials for running code tests](#acquire-taskcluster-credentials-for-running-code-tests)
      * [Option 1](#option-1)
      * [Option 2](#option-2)
   * [Set up your env](#set-up-your-env)
   * [Start the generic worker](#start-the-generic-worker)
   * [Create a test job](#create-a-test-job)
   * [Run the generic worker test suite](#run-the-generic-worker-test-suite)
   * [Making a new generic worker release](#making-a-new-generic-worker-release)
   * [Creating and updating worker types](#creating-and-updating-worker-types)
   * [Release notes](#release-notes)



# Introdution

Generic worker is a native Windows/Linux/macOS program for executing
taskcluster tasks. It communicates with the taskcluster Queue as per the
[Queue-Worker Interaction
specification](https://docs.taskcluster.net/docs/reference/platform/taskcluster-queue/docs/worker-interaction).
It is shipped as a statically linked system-native executable. It is written in
go (golang).

## Imperative task payloads

Generic worker allows you to execute arbitrary commands in a task.

If you wish to only run trusted code against
input parameters passed in task payloads, see:
* [scriptworker](https://github.com/mozilla-releng/scriptworker)

If you are looking to isolate your tasks inside docker containers, see:
* [docker-worker](https://github.com/taskcluster/docker-worker)

Please note docker support is coming to generic-worker in [PR
62](https://github.com/taskcluster/generic-worker/pull/62).

# Sandboxing

It is important that tasks run in a sandbox in order to that they are as
reproducible as possible, and are not inadvertently affected by previous tasks
that may have run on the same environment. Different operating systems provide
different sandboxing mechanisms, and therefore the approach used by
generic-worker is platform-dependent.

## Windows

On Windows, `generic-worker.exe` runs in a [Windows
Service](https://docs.microsoft.com/en-us/windows/desktop/services/services)
under the
[LocalSystem](https://docs.microsoft.com/en-us/windows/desktop/services/localsystem-account)
account.

The worker creates a unique Operating System user to sandbox the activity of
the task.

All task commands run as the task user. After the task has completed, the user
is deleted, together with any files it has created.

In order for tasks to have access to a graphical logon session, the host is
configured to logon on boot as the new task user, and the
machine is rebooted.

By default the generated users are standard (non-admin) OS users.

### Task user lifecycle
The worker configures the machine to automatically log
in as the newly created task user, and then triggers the machine to reboot.
Once the machine reboots, the worker running in the Windows Service waits until
it detects that the Operating System [winlogon
module](https://docs.microsoft.com/en-us/windows/desktop/secauthn/winlogon-and-credential-providers)
has completed the interactive logon of the task user. At this point it polls
the taskcluster Queue to fetch a task to execute, and when it is given one, it
executes this task in the interactive logon session of the logged-in user,
running processes using the auth token obtained from the interactive desktop
session.

After the task completes, the home directory of the task user, and the task
directory (if different to the home directory of the task user) are erased, a
new task user is created, the machine is rebooted, and the former task user is
purged.

In the same way that a guest account allows an untrusted user to temporarily
use a machine without impacting the rest of the machine, the generic worker
allows tasks to run on the host without having permanent affect.  After the
task has completed, all trace of the changes made by the task user should be
gone, and the machine's state should be reset to the state it had before the
task was run. If the host environment is sufficiently locked down, the task
user should not have been able to apply any state-change to the host
environment. Please note that the worker has limited control to affect
system-wide policy, so for example if a host allows arbtirary users to write to
a system folder location, the worker is not able to prevent a task doing so.
Therefore it is up to the machine provider to ensure that the host is
sufficiently locked-down. Host environments for long-lived workers that are to
run untrusted tasks should be secured carefully, to prevent that tasks may
interfere with system state or persist changes across task runs that may affect
the reproducibility of a task, or worse, introduce a security vulnerability.






## Linux

There is no native sandbox support currently on Linux. Currently the worker
will execute tasks as the same user that the worker runs as. Use at your own
risk!

Work is [underway](https://github.com/taskcluster/generic-worker/pull/62) to
provide support for running generic-worker tasks inside a docker container
isolated from the host environment. However until this work is complete, please
see [docker-worker](https://github.com/taskcluster/docker-worker) for achieving
this.

We may, at some point, provide OS-user sandboxing, akin to the Windows
implementation.


## macOS

There is no native sandbox support currently on macOS. Currently the worker
will execute tasks as the same user that the worker runs as. Use at your own
risk!

We intend to provide OS-user sandboxing, akin to the Windows implementation, at
some point in the future.

# Operating System integration

## Windows

### Creating a task user

The generic-worker creates non-privileged task users, with username
`task_<current-unix-timestamp>` and a random password. Task users are created
with the command:

```
net user "<username>" "<userpassword>" /add /expires:never /passwordchg:no /y
```

See [net
user](https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-xp/bb490718%28v%253dtechnet.10%29)
for more information.

### Setting Known Folder locations

It may be desirable for the task directory to be in a different location to the
home directory of the generated user. The location of the home directory is
determined based on system settings defined at installation time of the
operating system, and therefore may not be ideal, especially if the host image
is provided by an external party.

For example, perhaps the user home directory is `C:\Users\task_<timestamp>` but
for performance reasons, we wish the task directory to be located on a
different physical drive at `Z:\task_<timestamp>`.

It is possible to configure the location for the task directories (in the above
case, `Z:\`) via the `tasksDir` property of the generic-worker configuration
file. However, this would not affect the location of the [AppData
folder](https://www.howtogeek.com/318177/what-is-the-appdata-folder-in-windows/)
used by Windows applications, which would still be located under the user's
home directory.

Since it is usually preferable for all user data to be written
to the the task directory, and it isn't trivial to move the user home directory
to an alternative location after the operating system has already been
installed, the worker configures the user account to store the AppData folder
under the task directory.

It does this as follows:

1) Calling
[LogonUserW](https://docs.microsoft.com/en-us/windows/desktop/api/winbase/nf-winbase-logonuserw)
to get a logon handle for the new user.

2) Calling
[LoadUserProfileW](https://docs.microsoft.com/en-us/windows/desktop/api/userenv/nf-userenv-loaduserprofilew)
to load the user profile.

3) Calling
[SHSetKnownFolderPath](https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/nf-shlobj_core-shsetknownfolderpath)
with `KNOWNFOLDERID`
[FOLDERID_RoamingAppData](https://docs.microsoft.com/en-us/windows/desktop/shell/knownfolderid)
to set the location of `AppData\Roaming` to under the task directory.

4) Calling
[SHGetKnownFolderPath](https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/nf-shlobj_core-shgetknownfolderpath)
with
[KF_FLAG_CREATE](https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/ne-shlobj_core-known_folder_flag)
in order to create `AppData\Roaming` folder.

5) Calling [CoTaskMemFree](CoTaskMemFree) to release resources from
`SHGETKnownFolderPath` call in step 4.

6) Calling
[SHSetKnownFolderPath](https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/nf-shlobj_core-shsetknownfolderpath)
with `KNOWNFOLDERID`
[FOLDERID_LocalAppData](https://docs.microsoft.com/en-us/windows/desktop/shell/knownfolderid)
to set the location of `AppData\Local` to under the task directory.

7) Calling
[SHGetKnownFolderPath](https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/nf-shlobj_core-shgetknownfolderpath)
with
[KF_FLAG_CREATE](https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/ne-shlobj_core-known_folder_flag)
in order to create `AppData\Local` folder.

8) Calling [CoTaskMemFree](CoTaskMemFree) to release resources from
`SHGETKnownFolderPath` call in step 7.

9) Calling
[UnloadUserProfile](https://docs.microsoft.com/en-us/windows/desktop/api/userenv/nf-userenv-unloaduserprofile)
to release resources from `LoadUserProfileW` call in step 2.

10) Calling
[CloseHandle](https://msdn.microsoft.com/en-us/library/windows/desktop/ms724211%28v=vs.85%29.aspx?f=255&MSPPError=-2147217396)
to release resources from `LogonUserW` call in step 1.


### Configuring auto-logon of task user

After the task user has been created, the Windows registry is updated so that
after rebooting, the task user will be automatically logged in.

This is achieved by configuring the registry key values:

* `\HKEY_LOCAL_MACHINE\Software\Microsoft\Windows NT\CurrentVersion\Winlogon\AutoAdminLogon = 1`
* `\HKEY_LOCAL_MACHINE\Software\Microsoft\Windows NT\CurrentVersion\Winlogon\DefaultUserName = <task user username>`
* `\HKEY_LOCAL_MACHINE\Software\Microsoft\Windows NT\CurrentVersion\Winlogon\DefaultPassword = <task user password>`

See [Automatic
Logon](https://docs.microsoft.com/en-us/windows/desktop/secauthn/msgina-dll-features)
for more detailed information about these settings.

### Rebooting

Rebooting is achieved by executing:

```
C:\Windows\System32\shutdown.exe /r /t 3 /c "generic-worker requested reboot"
```

Please note, automatic reboots can be disabled (see `generic-worker --help` for
more information).

### Executing task commands

The Windows Command Shell does not have a setting to enable exit-on-fail
semantics. Execution of a batch script continues if a command fails. To cause
a batch script to exit after a failed command, the exit code of every command
needs to be checked, or commands need to be chained together with `&&`.

Since this is cumbersome or error-prone, generic-worker accepts task payloads
with multiple commands. It will execute them in sequence with exit-on-fail
semantics. Each command is implicitly executed with `cmd.exe`, which means that
commands may contain any valid [command shell syntax](https://ss64.com/nt/).

Other workers (such as docker worker) accept only a single task command. If a
task wishes to execute multiple commands, it will usually specify a single
shell command to execute them. This approach works well when the shell supports
exit-on-fail semantics, but not so well when it doesn't, which is why a
different approach was chosen for generic-worker.

Generic worker generates a wrapper batch script for each command it runs, which
initialises environment variables, sets the working directory, executes the
task command, and then if more commands are to follow, captures the working
directory and environment variables for the next command.



# Payload format

Each taskcluster task definition contains a top level property `payload` which
is a json object. The format of this object is specific to the worker
implementation. For generic-worker, this is then also further specific to the
platform (Linux/Windows/macOS).

The per-platform payload formats are described in json schema, and can be found
in the top level `schemas` subdirectory of this repository. These schemas are
also published to the [generic-worker
page](https://docs.taskcluster.net/docs/reference/workers/generic-worker/docs/payload)
of the taskcluster docs site.

# Redeployability

# Integrating with AWS / GCE

# Config bootstrapping

# Bring your own worker

This section explains how to configure and run your own generic-worker workers
to talk to an existing taskcluster deployment.

## Windows
### Installing
## Mac
### Installing

There currently is no `install` target for macOS, like there is for Windows.

For our own dedicated macOS workers, we install generic-worker using [this
puppet
module](https://wiki.mozilla.org/ReleaseEngineering/PuppetAgain/Modules/generic_worker).

You can install generic-worker as a Launch Agent as follows:

1) Create a regular unprivileged user account on your Mac to run the worker
(e.g. with name `genericworker`) and log into that user account.

2) Download or build generic-worker, so that you have a native darwin binary,
move it to `/usr/local/bin/generic-worker`, and make sure it is executable for
your new user (`chmod u+x /usr/local/bin/generic-worker`).

3) Create a signing key in the user home directory by running:

```
/usr/local/bin/generic-worker new-openpgp-keypair --file .signingkey
```

3) Create `/Library/LaunchAgents/net.generic.worker.plist` with content:

```
<%# This Source Code Form is subject to the terms of the Mozilla Public
<%# License, v. 2.0. If a copy of the MPL was not distributed with this
<%# file, You can obtain one at http://mozilla.org/MPL/2.0/. -%>
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>net.generic.worker</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/generic-worker</string>
        <string>run</string>
        <string>--config</string>
        <string>/etc/generic-worker.config</string>
    </array>
    <key>WorkingDirectory</key>
    <string><-YOUR-NEW-USER-HOME-DIRECTORY-></string>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

4) Create `/etc/generic-worker.config` with content:

```
{
  "accessToken": "<-YOUR-TASKCLUSTER-ACCESS-TOKEN->",
  "clientId": "<-YOUR-TASKCLUSTER-CLIENT-ID->",
  "idleTimeoutSecs": 0,
  "livelogSecret": "<-RANDOM-SHORT-STRING-HERE->",
  "provisionerId": "<-YOUR-PROVISIONER-ID->",
  "publicIP": "<-MAKE-UP-AN-IPv4-ADDRESS-IF-YOU-DON'T-HAVE-ONE->",
  "signingKeyLocation": ".signingkey",
  "tasksDir": "tasks",
  "workerGroup": "<-CHOOSE-A-WORKER-GROUP->",
  "workerId": "<-CHOOSE-A-WORKER-ID->",
  "workerType": "<-YOUR-WORKER-TYPE->",
  "workerTypeMetadata": {
	<--- add a json blob here with information about you, how you set up the
         worker type, etc, so people know how it is configured and maintained,
         and who to go to in case of problems --->
  }
}
```

## Linux - Docker
## Linux - Native
### Installing
# Administrative tools
## Displaying workers
# Worker Type Host Definitions
## Updating existing definitions
## Modifying definitions
## Creating your own AWS workers outside of this repo
## Puppet
# Developing Generic Worker
## Fetching source
## Credentials
## Running unit tests
## Writing unit tests
## Including bug numbers in comments

# Releasing Generic Worker
## Release script
## Publishing schemas
## Testing in Staging
## Rolling out to Production
## Writing release notes (README.md, release page, ...)
# Repository layout

```
├── aws
│   ├── cmd
│   │   ├── download-aws-worker-type-definitions
│   │   ├── gw-workers
│   │   ├── update-ssl-creds
│   │   └── update-worker-type
│   ├── scripts
│   └── update-worker-types
├── cmd
│   ├── generic-worker
│   ├── gw-codegen
│   ├── inspect-worker-types
│   ├── list-worker-types
│   └── yamltojson
├── docs
├── lib
├── mozilla
│   ├── OpenCloudConfig
│   │   ├── occ-workers
│   │   ├── refresh-gw-configs
│   │   └── transform-occ
│   ├── gecko
│   ├── nss
│   └── worker-type-host-definitions
│       └── aws-provisioner-v1
│           ├── <worker type>
│           ├── <worker type>
│           └── ...
├── schemas
└── scripts
```


# Downloading generic-worker binary release

* Download the latest release for your platform from https://github.com/taskcluster/generic-worker/releases
* Download the latest release of livelog for your platform from https://github.com/taskcluster/livelog/releases
* For darwin/linux, make the binaries executable: `chmod a+x {generic-worker,livelog}*`

# Building generic-worker from source

If you prefer not to use a prepackaged binary, or want to have the latest unreleased version from the development head:

* Head over to https://golang.org/dl/ and follow the instructions for your platform. __Note, go 1.8 or higher is required__.
* Run `./build.sh` (macOS/Linux) or `.\build.cmd` (Windows)

All being well, the binaries will be built under `$(go env GOPATH)/bin`.



# Acquire taskcluster credentials for running code tests

There are two alternative mechanisms to acquire the scopes you need.

## Option 1

This method works if you log into Taskcluster via mozillians, *or* you log into
taskcluster via LDAP *using the same email address as your mozillians account*,
*or* if you do not currently have a mozillians account but would like to create
one.

* Sign up for a [Mozillians account](https://mozillians.org/en-US/) (if you do not already have one)
* Request membership of the [taskcluster-contributors](https://mozillians.org/en-US/group/taskcluster-contributors/) mozillians group

## Option 2

This method is for those who wish not to create a mozillians account, but
already authenticate into taskcluster via some other means, or have a
mozillians account but it is registered to a different email address than the
one they use to log into Taskcluster with (e.g. via LDAP integration).

* Request the scope `assume:project:taskcluster:generic-worker-tester` to be
  granted to you via a [bugzilla
  request](https://bugzilla.mozilla.org/enter_bug.cgi?product=Taskcluster&component=Service%20Request),
  including your [currently active `ClientId`](https://tools.taskcluster.net/credentials/)
  in the bug description. From the ClientId, we will be able to work out which role to assign the scope
  to, in order that you acquire the scope with the client you log into Taskcluster tools site with.

Once you have been granted the above scope:

* If you are signed into tools.taskcluster.net already, **sign out**
* Sign into [tools.taskcluster.net](https://tools.taskcluster.net/) using either your new Mozillians account, _or_ your LDAP account **if it uses the same email address as your Mozillians account**
* Check that a role or client of yours appears in [this list](https://tools.taskcluster.net/auth/scopes/assume%3Aproject%3Ataskcluster%3Ageneric-worker-tester)
* Create a permanent client (taskcluster credentials) for yourself in the [Client Manager](https://tools.taskcluster.net/auth/clients/) granting it the single scope `assume:project:taskcluster:generic-worker-tester`

# Set up your env

* Generate a GPG key pair with `generic-worker new-openpgp-keypair --file <file>` where `file` is where you want the generated GPG private key to be written to
* Create a generic worker configuration file somewhere, with the following content:

```
{
    "accessToken":                "<access token of your permanent credentials>",
    "certificate":                "",
    "clientId":                   "<client ID of your permanent credentials>",
    "livelogSecret":              "<anything you like>",
    "provisionerId":              "test-provisioner",
    "publicIP":                   "<ideally an IP address of one of your network interfaces>",
    "signingKeyLocation":         "<file location you wrote gpg private key to>",
    "workerGroup":                "test-worker-group",
    "workerId":                   "test-worker-id",
    "workerType":                 "<a unique name that only you will use for your test worker(s)>"
}
```

To see a full description of all the config options available to you, run `generic-worker --help`:

```
generic-worker 11.0.1

generic-worker is a taskcluster worker that can run on any platform that supports go (golang).
See http://taskcluster.github.io/generic-worker/ for more details. Essentially, the worker is
the taskcluster component that executes tasks. It requests tasks from the taskcluster queue,
and reports back results to the queue.

  Usage:
    generic-worker run                      [--config         CONFIG-FILE]
                                            [--configure-for-aws]
    generic-worker install service          [--nssm           NSSM-EXE]
                                            [--service-name   SERVICE-NAME]
                                            [--config         CONFIG-FILE]
                                            [--configure-for-aws]
    generic-worker show-payload-schema
    generic-worker new-openpgp-keypair      --file PRIVATE-KEY-FILE
    generic-worker grant-winsta-access      --sid SID
    generic-worker --help
    generic-worker --version

  Targets:
    run                                     Runs the generic-worker.
    show-payload-schema                     Each taskcluster task defines a payload to be
                                            interpreted by the worker that executes it. This
                                            payload is validated against a json schema baked
                                            into the release. This option outputs the json
                                            schema used in this version of the generic
                                            worker.
    install service                         This will install the generic worker as a
                                            Windows service running under the Local System
                                            account. This is the preferred way to run the
                                            worker under Windows. Note, the service will
                                            be configured to start automatically. If you
                                            wish the service only to run when certain
                                            preconditions have been met, it is recommended
                                            to disable the automatic start of the service,
                                            after you have installed the service, and
                                            instead explicitly start the service when the
                                            preconditions have been met.
    new-openpgp-keypair                     This will generate a fresh, new OpenPGP
                                            compliant private/public key pair. The public
                                            key will be written to stdout and the private
                                            key will be written to the specified file.
    grant-winsta-access                     Windows only. Used internally by generic-
                                            worker to grant a logon SID full control of the
                                            interactive windows station and desktop.

  Options:
    --config CONFIG-FILE                    Json configuration file to use. See
                                            configuration section below to see what this
                                            file should contain. When calling the install
                                            target, this is the config file that the
                                            installation should use, rather than the config
                                            to use during install.
                                            [default: generic-worker.config]
    --configure-for-aws                     Use this option when installing or running a worker
                                            that is spawned by the AWS provisioner. It will cause
                                            the worker to query the EC2 metadata service when it
                                            is run, in order to retrieve data that will allow it
                                            to self-configure, based on AWS metadata, information
                                            from the provisioner, and the worker type definition
                                            that the provisioner holds for the worker type.
    --nssm NSSM-EXE                         The full path to nssm.exe to use for installing
                                            the service.
                                            [default: C:\nssm-2.24\win64\nssm.exe]
    --service-name SERVICE-NAME             The name that the Windows service should be
                                            installed under. [default: Generic Worker]
    --file PRIVATE-KEY-FILE                 The path to the file to write the private key
                                            to. The parent directory must already exist.
                                            If the file exists it will be overwritten,
                                            otherwise it will be created.
    --sid SID                               A SID to be granted full control of the
                                            interactive windows station and desktop, for
                                            example: 'S-1-5-5-0-41431533'.
    --help                                  Display this help text.
    --version                               The release version of the generic-worker.


  Configuring the generic worker:

    The configuration file for the generic worker is specified with -c|--config CONFIG-FILE
    as described above. Its format is a json dictionary of name/value pairs.

        ** REQUIRED ** properties
        =========================

          accessToken                       Taskcluster access token used by generic worker
                                            to talk to taskcluster queue.
          clientId                          Taskcluster client id used by generic worker to
                                            talk to taskcluster queue.
          livelogSecret                     This should match the secret used by the
                                            stateless dns server; see
                                            https://github.com/taskcluster/stateless-dns-server
          publicIP                          The IP address for clients to be directed to
                                            for serving live logs; see
                                            https://github.com/taskcluster/livelog and
                                            https://github.com/taskcluster/stateless-dns-server
          signingKeyLocation                The PGP signing key for signing artifacts with.
          workerId                          A name to uniquely identify your worker.
          workerType                        This should match a worker_type managed by the
                                            provisioner you have specified.

        ** OPTIONAL ** properties
        =========================

          authBaseURL                       The base URL for API calls to the auth service.
          availabilityZone                  The EC2 availability zone of the worker.
          cachesDir                         The directory where task caches should be stored on
                                            the worker. The directory will be created if it does
                                            not exist. This may be a relative path to the
                                            current directory, or an absolute path.
                                            [default: caches]
          certificate                       Taskcluster certificate, when using temporary
                                            credentials only.
          checkForNewDeploymentEverySecs    The number of seconds between consecutive calls
                                            to the provisioner, to check if there has been a
                                            new deployment of the current worker type. If a
                                            new deployment is discovered, worker will shut
                                            down. See deploymentId property. [default: 1800]
          cleanUpTaskDirs                   Whether to delete the home directories of the task
                                            users after the task completes. Normally you would
                                            want to do this to avoid filling up disk space,
                                            but for one-off troubleshooting, it can be useful
                                            to (temporarily) leave home directories in place.
                                            Accepted values: true or false. [default: true]
          deploymentId                      If running with --configure-for-aws, then between
                                            tasks, at a chosen maximum frequency (see
                                            checkForNewDeploymentEverySecs property), the
                                            worker will query the provisioner to get the
                                            updated worker type definition. If the deploymentId
                                            in the config of the worker type definition is
                                            different to the worker's current deploymentId, the
                                            worker will shut itself down. See
                                            https://bugzil.la/1298010
          disableReboots                    If true, no system reboot will be initiated by
                                            generic-worker program, but it will still return
                                            with exit code 67 if the system needs rebooting.
                                            This allows custom logic to be executed before
                                            rebooting, by patching run-generic-worker.bat
                                            script to check for exit code 67, perform steps
                                            (such as formatting a hard drive) and then
                                            rebooting in the run-generic-worker.bat script.
                                            [default: false]
          downloadsDir                      The directory to cache downloaded files for
                                            populating preloaded caches and readonly mounts. The
                                            directory will be created if it does not exist. This
                                            may be a relative path to the current directory, or
                                            an absolute path. [default: downloads]
          idleTimeoutSecs                   How many seconds to wait without getting a new
                                            task to perform, before the worker process exits.
                                            An integer, >= 0. A value of 0 means "never reach
                                            the idle state" - i.e. continue running
                                            indefinitely. See also shutdownMachineOnIdle.
                                            [default: 0]
          instanceID                        The EC2 instance ID of the worker.
          instanceType                      The EC2 instance Type of the worker.
          livelogCertificate                SSL certificate to be used by livelog for hosting
                                            logs over https. If not set, http will be used.
          livelogExecutable                 Filepath of LiveLog executable to use; see
                                            https://github.com/taskcluster/livelog
                                            [default: livelog]
          livelogGETPort                    Port number for livelog HTTP GET requests.
                                            [default: 60023]
          livelogKey                        SSL key to be used by livelog for hosting logs
                                            over https. If not set, http will be used.
          livelogPUTPort                    Port number for livelog HTTP PUT requests.
                                            [default: 60022]
          numberOfTasksToRun                If zero, run tasks indefinitely. Otherwise, after
                                            this many tasks, exit. [default: 0]
          privateIP                         The private IP of the worker, used by chain of trust.
          provisionerBaseURL                The base URL for API calls to the provisioner in
                                            order to determine if there is a new deploymentId.
          provisionerId                     The taskcluster provisioner which is taking care
                                            of provisioning environments with generic-worker
                                            running on them. [default: test-provisioner]
          purgeCacheBaseURL                 The base URL for API calls to the purge cache
                                            service.
          queueBaseURL                      The base URL for API calls to the queue service.
          region                            The EC2 region of the worker.
          requiredDiskSpaceMegabytes        The garbage collector will ensure at least this
                                            number of megabytes of disk space are available
                                            when each task starts. If it cannot free enough
                                            disk space, the worker will shut itself down.
                                            [default: 10240]
          runAfterUserCreation              A string, that if non-empty, will be treated as a
                                            command to be executed as the newly generated task
                                            user, after the user has been created, the machine
                                            has rebooted and the user has logged in, but before
                                            a task is run as that user. This is a way to
                                            provide generic user initialisation logic that
                                            should apply to all generated users (and thus all
                                            tasks) and be run as the task user itself. This
                                            option does *not* support running a command as
                                            Administrator.
          runTasksAsCurrentUser             If true, users will not be created for tasks, but
                                            the current OS user will be used. [default: true]
          sentryProject                     The project name used in https://sentry.io for
                                            reporting worker crashes. Permission to publish
                                            crash reports is granted via the scope
                                            auth:sentry:<sentryProject>. If the taskcluster
                                            client (see clientId property above) does not
                                            posses this scope, no crash reports will be sent.
                                            Similarly, if this property is not specified or
                                            is the empty string, no reports will be sent.
          shutdownMachineOnIdle             If true, when the worker is deemed to have been
                                            idle for enough time (see idleTimeoutSecs) the
                                            worker will issue an OS shutdown command. If false,
                                            the worker process will simply terminate, but the
                                            machine will not be shut down. [default: false]
          shutdownMachineOnInternalError    If true, if the worker encounters an unrecoverable
                                            error (such as not being able to write to a
                                            required file) it will shutdown the host
                                            computer. Note this is generally only desired
                                            for machines running in production, such as on AWS
                                            EC2 spot instances. Use with caution!
                                            [default: false]
          subdomain                         Subdomain to use in stateless dns name for live
                                            logs; see
                                            https://github.com/taskcluster/stateless-dns-server
                                            [default: taskcluster-worker.net]
          taskclusterProxyExecutable        Filepath of taskcluster-proxy executable to use; see
                                            https://github.com/taskcluster/taskcluster-proxy
                                            [default: taskcluster-proxy]
          taskclusterProxyPort              Port number for taskcluster-proxy HTTP requests.
                                            [default: 80]
          tasksDir                          The location where task directories should be
                                            created on the worker. [default: /Users]
          workerGroup                       Typically this would be an aws region - an
                                            identifier to uniquely identify which pool of
                                            workers this worker logically belongs to.
                                            [default: test-worker-group]
          workerTypeMetaData                This arbitrary json blob will be included at the
                                            top of each task log. Providing information here,
                                            such as a URL to the code/config used to set up the
                                            worker type will mean that people running tasks on
                                            the worker type will have more information about how
                                            it was set up (for example what has been installed on
                                            the machine).

    If an optional config setting is not provided in the json configuration file, the
    default will be taken (defaults documented above).

    If no value can be determined for a required config setting, the generic-worker will
    exit with a failure message.

  Exit Codes:

    0      Tasks completed successfully; no more tasks to run (see config setting
           numberOfTasksToRun).
    64     Not able to load specified generic-worker config file.
    65     Not able to install generic-worker on the system.
    66     Not able to create an OpenPGP key pair.
    67     A task user has been created, and the generic-worker needs to reboot in order
           to log on as the new task user. Note, the reboot happens automatically unless
           config setting disableReboots is set to true - in either code this exit code will
           be issued.
    68     The generic-worker hit its idle timeout limit (see config settings idleTimeoutSecs
           and shutdownMachineOnIdle).
    69     Worker panic - either a worker bug, or the environment is not suitable for running
           a task, e.g. a file cannot be written to the file system, or something else did
           not work that was required in order to execute a task. See config setting
           shutdownMachineOnInternalError.
    70     A new deploymentId has been issued in the AWS worker type configuration, meaning
           this worker environment is no longer up-to-date. Typcially workers should
           terminate.
    71     The worker was terminated via an interrupt signal (e.g. Ctrl-C pressed).
    72     The worker is running on spot infrastructure in AWS EC2 and has been served a
           spot termination notice, and therefore has shut down.
    73     The config provided to the worker is invalid.
    74     Could not grant provided SID full control of interactive windows stations and
           desktop.
```

# Start the generic worker

Simply run:

```
generic-worker run --config <config file>
```

where `<config file>` is the generic worker config file you created above.

# Create a test job

Go to https://tools.taskcluster.net/task-creator/ and create a task to run on your generic worker.

Use [this example](worker_types/win2012r2/task-definition.json) as a template, but make sure to edit `provisionerId` and `workerType` values so that they match what you set in your config file.

Don't forget to submit the task by clicking the *Create Task* icon.

If all is well, your local generic worker should pick up the job you submit, run it, and report back status.

# Run the generic worker test suite

For this you need to have the source files (you cannot run the tests from the binary package).

Then cd into the source directory, and run:

```
go test -v ./...
```

# Making a new generic worker release

Run the `release.sh` script like so:

```
$ ./release.sh 11.0.1
```

This will perform some checks, tag the repo, push the tag to github, which will then trigger travis-ci to run tests, and publish the new release.

# Creating and updating worker types

See [worker_types README.md](https://github.com/taskcluster/generic-worker/blob/master/worker_types/README.md).

# Release notes

### In v10.11.3 since v10.11.2

* [Bug 1480412 - allow empty osGroups list on non-Windows platforms](https://bugzil.la/1480412#c10)

### In v10.11.2 since v10.11.1

* [Bug 1475689 - osx generic-worker not rebooting after some failed jobs](https://bugzil.la/1475689)

### In v10.11.1 since v10.11.1alpha1

* [Bug 1439588 - Add feature to support running Windows tasks in an elevated process (Administrator)](https://bugzil.la/1439588)

### In v10.10.0 since v10.9.0

* [Bug 1469402 - Support for onExitStatus on generic-worker](https://bugzil.la/1469402)

### In v10.8.5 since v10.8.4

* [Bug 1468155 - Enforce that tasks depend on the tasks whose content they mount](https://bugzil.la/1468155)

### In v10.8.4 since v10.8.3

* [Bug 1433854 - clean up after tasks on windows test workers](https://bugzil.la/1433854)
* [Bug 1465479 - Don't block task abortion waiting for cmd.Wait() to complete](https://bugzil.la/1465479)
* [Bug 1466803 - Could not copy from backing log to livelog: io: read/write on closed pipe](https://bugzil.la/1466803)

### In v10.8.2 since v10.8.1

* [Bug 1462369 - Kill entire process tree when aborting a task](https://bugzil.la/1462369)

### In v10.8.0 since v10.7.12

* [Bug 1459376 - Log information about files downloaded via "mounts"](https://bugzil.la/1459376)

### In v10.7.12 since v10.7.11

* [Bug 1452095 - Upgrade mac taskcluster workers to generic-worker 10.8.4](https://bugzil.la/1452095)
* [Bug 1458873 - Process termination when aborting task not always successful on Windows](https://bugzil.la/1458873)

### In v10.7.11 since v10.7.10

* [Bug 1456357 - Intermittent [taskcluster:error] Get http://schemas.taskcluster.net/generic-worker/v1/payload.json: dial tcp 52.84.128.102:80: i/o timeout](https://bugzil.la/1456357)

### In v10.7.10 since v10.7.9

* [Bug 1444118 - Better log message if chain of trust key is not valid](https://bugzil.la/1444118)

### In v10.7.6 since v10.7.5

* [Bug 1447265 - Use go 1.10 os/exec package for running processes with CreateProcessAsUser on Windows](https://bugzil.la/1447265)

### In v10.7.3 since v10.7.2

* [Bug 1180187 - generic-worker: listen for and handle worker shutdown](https://bugzil.la/1180187)

### In v10.6.1 since v10.6.0

* [Bug 1443595 - github binary downloads are broken in occ due to the tls upgrade](https://bugzil.la/1443595)

### In v10.6.0 since v10.5.1

* [Bug 1358545 - [generic-worker] startup tests for CoT privkey protection](https://bugzil.la/1358545)
* [Bug 1439517 - generic-worker: support taskcluster-proxy](https://bugzil.la/1439517)
* [Bug 1441482 - generic-worker should free up disk space *before* claiming a task](https://bugzil.la/1441482)

### In v10.5.1 since v10.5.0

* [Bug 1333957 - Make "Aborting task - max run time exceeded!" a Treeherder-parseable message](https://bugzil.la/1333957)

### In v10.5.0 since v10.5.0alpha4

* [Bug 1172273 - generic-worker (windows): RDP into task users](https://bugzil.la/1172273)
* [Bug 1429370 - Changing PR title causes new tasks to be triggered](https://bugzil.la/1429370)

### In v10.4.1 since v10.4.0

* [Bug 1424986 - No attempt in logs to reclaim task](https://bugzil.la/1424986)
* [Bug 1425438 - Idle time inaccurate when (and after) computer sleeps/hibernates/is not being watched](https://bugzil.la/1425438)

### In v10.4.0 since v10.3.1

* [Bug 1423215 - Uploaded gzipped job artifacts (such as runnable-jobs.json.gz) have incorrect Content-Encoding/Type](https://bugzil.la/1423215)

### In v10.3.0 since v10.2.3

* [Bug 1415088 - Include generic worker version number when reporting crashes to sentry](https://bugzil.la/1415088)

### In v10.2.3 since v10.2.2

* [Bug 1397373 - Move superseding docs out of docker-worker](https://bugzil.la/1397373)
* [Bug 1401007 - Ensure generic worker runs on Windows Server 2016](https://bugzil.la/1401007)
* [Bug 1402152 - Use temporary credentials from claimWork, reclaimTask in reclaimTask, createArtifact, reportCompleted](https://bugzil.la/1402152)

### In v10.2.2 since v10.2.1

* [Bug 1382204 - Enable coalescing for scm level 2,3 test tasks on macOS and win10 gpu and linux](https://bugzil.la/1382204)

### In v10.2.1 since v10.2.0

* [Bug 1394557 - Intermittent "400 Bad Request" errors when uploading to the TC queue causing Windows job failures](https://bugzil.la/1394557)

### In v10.2.0 since v10.1.8

* [Bug 1383024 - generic-worker: Implement coalescing/superseding](https://bugzil.la/1383024)

### In v10.1.8 since v10.1.7

* [Bug 1387015 - Python wheel artifact should not be gzipped](https://bugzil.la/1387015)

### In v10.1.7 since v10.1.6

* [Bug 1385870 - generic-worker: Do not gzip content encode artifacts with .xz extension](https://bugzil.la/1385870)

### In v10.1.6 since v10.1.5

* [Bug 1381801 - Task payload timestamps have inconsistent precision with top level task timestamps](https://bugzil.la/1381801)

### In v10.1.5 since v10.1.4

* [Bug 1360198 - Don't explicitly set artifact expiry for generic-worker tasks, if it is just task expiry](https://bugzil.la/1360198)

### In v10.1.4 since v10.1.3

* [Bug 1380978 - generic-worker: chain of trust artifacts should be indexed by artifact name, not artifact path](https://bugzil.la/1380978)

### In v10.0.5 since v10.0.4

* [Bug 1372210 - CoT on generic worker for windows uploads to 'public/logs/' when it should upload to 'public/'](https://bugzil.la/1372210)

### In v8.5.0 since v8.4.1

* [Bug 1360539 - generic-worker 8.3.0 to 8.4.1 causing process failures on win2012r2 worker types (runTasksAsCurrentUser: false)](https://bugzil.la/1360539)

### In v8.3.0 since v8.2.0

* [Bug 1347956 - Some public-artifacts.taskcluster.net files are not served gzipped](https://bugzil.la/1347956)

### In v8.2.0 since v8.1.1

* [Bug 1356800 - [generic-worker] Handle some errors during artifact uploading more gracefully](https://bugzil.la/1356800)

### In v8.1.0 since v8.0.1

* [Bug 1352457 - Support artifact name in task payload](https://bugzil.la/1352457)

### In v8.0.1 since v8.0.0

* [Bug 1337132 - 64-bit Windows static builds are very frequently timing out since the worker change from bug 1336948 happened](https://bugzil.la/1337132)

### In v7.2.13 since v7.2.12

* [Bug 1261188 - test_Edge_availability.js fails on Win10 in automation](https://bugzil.la/1261188)

### In v7.2.6 since v7.2.5

* [Bug 1329617 - generic-worker 7.2.5 stops claiming tasks after a task it is running gets cancelled](https://bugzil.la/1329617)

### In v7.2.2 since v7.2.1

* [Bug 1323827 - [generic-worker] Worker tries reclaiming resolved task](https://bugzil.la/1323827)

### In v7.1.1 since v7.1.0

* [Bug 1312383 - taskcluster windows 7 ec2 instances underperform](https://bugzil.la/1312383)

### In v7.1.0 since v7.0.3alpha1

* [Bug 1307204 - Convert gecko-L-b-win2012 workers to c4.4xlarge](https://bugzil.la/1307204)

### In v7.0.3alpha1 since v7.0.2alpha1

* [Bug 1303455 - TC Windows tests run without JOB_OBJECT_LIMIT_BREAKAWAY_OK](https://bugzil.la/1303455)

### In v6.1.0 since v6.1.0alpha1

* [Bug 1298010 - update Generic Worker to terminate host instance if the worker type definition has changed since the instance was started](https://bugzil.la/1298010)

### In v6.0.0 since v5.4.0

* [Bug 1306988 - GenericWorker task user groups should be configurable or task dependent](https://bugzil.la/1306988)
* [Bug 1307383 - win2012r2 hung / didn't run correctly](https://bugzil.la/1307383)

### In v5.4.0 since v5.3.1

* [Bug 1182451 - preloaded writable directory caches / empty writable directory caches / preloaded readonly directory mounts / preloaded readonly files](https://bugzil.la/1182451)
* [Bug 1305048 - Add node/npm to the generic "win2012r2" worker](https://bugzil.la/1305048)

### In v5.3.0 since v5.2.0

* [Bug 1287112 - enable chain-of-trust artifact generation in generic-worker](https://bugzil.la/1287112)

### In v5.1.0 since v5.0.3

* [Bug 1291249 - Logs uploaded by generic-worker to public-artifacts.taskcluster.net aren't using gzip](https://bugzil.la/1291249)

### In v4.0.0alpha1 since v3.0.0alpha1

* [Bug 1181524 - generic-worker: Reject tasks w. malformed-payload if artifact.expires x3c task.deadline](https://bugzil.la/1181524)
* [Bug 1191524 - [Browser] Suggestions in youtube search bar overlap with icon.](https://bugzil.la/1191524)
* [Bug 1285197 - package github.com/taskcluster/taskcluster-client-go: no buildable Go source files in x3cGOPATHx3e/src/github.com/taskcluster/taskcluster-client-go](https://bugzil.la/1285197)

### In v3.0.0alpha1 since v2.1.0

* [Bug 1279019 - Generic worker should log metadata about itself and the worker type it is running on](https://bugzil.la/1279019)

### In v2.0.0alpha44 since v2.0.0alpha43

* [Bug 1277568 - Generic worker live log artifacts are unreachable after task completes](https://bugzil.la/1277568)


# Further information

Please see:

* [Taskcluster Documentation](https://docs.taskcluster.net/)
* [Generic Worker presentations](https://docs.taskcluster.net/presentations) (focus on Windows platform)
* [Taskcluster Web Tools](https://tools.taskcluster.net/)
* [Generic Worker Open Bugs](https://bugzilla.mozilla.org/buglist.cgi?f1=product&resolution=---&o1=equals&o2=equals&query_format=advanced&f2=component&v1=Taskcluster&v2=Generic-Worker)

Useful information on win32 APIs:

* [Starting an Interactive Client Process in C++](https://msdn.microsoft.com/en-us/9e9ed9b7-ea23-4dec-8b92-a86aa81267ab?f=255&MSPPError=-2147217396)
* [Getting the Logon SID in C++](https://msdn.microsoft.com/en-us/aa446670?f=255&MSPPError=-2147217396)
* [Modifying the ACLs of an Object in C++](https://docs.microsoft.com/en-us/windows/desktop/secauthz/modifying-the-acls-of-an-object-in-c--)
* [Window Station Security and Access Rights](https://docs.microsoft.com/en-us/windows/desktop/winstation/window-station-security-and-access-rights)
