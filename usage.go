package main

import (
	"runtime"
	"strconv"
)

func usage(versionName string) string {
	return versionName + `

generic-worker is a taskcluster worker that can run on any platform that supports go (golang).
See http://taskcluster.github.io/generic-worker/ for more details. Essentially, the worker is
the taskcluster component that executes tasks. It requests tasks from the taskcluster queue,
and reports back results to the queue.

  Usage:
    generic-worker run                      [--config         CONFIG-FILE]
                                            [--configure-for-aws | --configure-for-gcp]
    generic-worker install service          [--nssm           NSSM-EXE]
                                            [--service-name   SERVICE-NAME]
                                            [--config         CONFIG-FILE]
                                            [--configure-for-aws | --configure-for-gcp]
    generic-worker show-payload-schema
    generic-worker new-ed25519-keypair      --file ED25519-PRIVATE-KEY-FILE
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
    new-ed25519-keypair                     This will generate a fresh, new ed25519
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
    --configure-for-gcp                     This will create the CONFIG-FILE for a GCP
                                            installation by querying the GCP environment
                                            and setting appropriate values.
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

        AUTH properties
        ===============

        There are three options for configuring taskcluster authentication:

        1) Permanent taskcluster credentials

          * accessToken, clientId, rootURL should be specified
          * certificate, proxyURL should be empty strings or unspecified

        2) Temporary taskcluster credentials

          * accessToken, clientId, certificate, rootURL should be specified
          * proxyURL should be empty strings or unspecified

        3) Authentication via a proxy

          * proxyURL, rootURL should be specified
          * accessToken, clientId, certificate should be empty strings or unspecified


          accessToken                       Taskcluster access token used by generic worker
                                            to talk to taskcluster queue.
          certificate                       Taskcluster certificate, when using temporary
                                            credentials only.
          clientId                          Taskcluster client ID used by generic worker to
                                            talk to taskcluster queue.
          proxyURL                          If running the worker _behind_ a taskcluster proxy,
                                            the taskcluster proxy url. Note this is mostly only
                                            useful for the generic-worker CI itself, but if
                                            desired, any worker can run behind a taskcluster
                                            proxy in order to keep taskcluster credentials
                                            away from the worker. If a proxyURL is specified,
                                            please note that clientId / accessToken / certificate
                                            should all be empty, but rootURL is still required.
          rootURL                           The root URL of the taskcluster deployment to which
                                            clientId and accessToken grant access. For example,
                                            'https://taskcluster.net'. Individual services can
                                            override this setting - see the *BaseURL settings.

        ** REQUIRED ** properties
        =========================

          ed25519SigningKeyLocation         The ed25519 signing key for signing artifacts with.
          publicIP                          The IP address for clients to be directed to
                                            for serving live logs; see
                                            https://github.com/taskcluster/livelog and
                                            https://github.com/taskcluster/stateless-dns-server
                                            Also used by chain of trust.
          workerId                          A name to uniquely identify your worker.
          workerType                        This should match a worker_type managed by the
                                            provisioner you have specified.

        ** OPTIONAL ** properties
        =========================

          authBaseURL                       The base URL for taskcluster auth API calls.
                                            If not provided, the base URL for API calls is
                                            instead derived from rootURL setting as follows:
                                              * https://auth.taskcluster.net/v1 for rootURL https://taskcluster.net
                                              * <rootURL>/api/auth/v1 for all other rootURLs
          availabilityZone                  The EC2 availability zone of the worker.
          cachesDir                         The directory where task caches should be stored on
                                            the worker. The directory will be created if it does
                                            not exist. This may be a relative path to the
                                            current directory, or an absolute path.
                                            [default: caches]
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
          instanceID                        The EC2 instance ID of the worker. Used by chain of trust.
          instanceType                      The EC2 instance Type of the worker. Used by chain of trust.
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
          livelogSecret                     This should match the secret used by the
                                            stateless dns server; see
                                            https://github.com/taskcluster/stateless-dns-server
                                            Optional if stateless DNS is not in use.
          numberOfTasksToRun                If zero, run tasks indefinitely. Otherwise, after
                                            this many tasks, exit. [default: 0]
          privateIP                         The private IP of the worker, used by chain of trust.
          provisionerBaseURL                The base URL for aws-provisioner API calls.
                                            If not provided, the base URL for API calls is
                                            instead derived from rootURL setting as follows:
                                              * https://aws-provisioner.taskcluster.net/v1 for rootURL https://taskcluster.net
                                              * <rootURL>/api/aws-provisioner/v1 for all other rootURLs
          provisionerId                     The taskcluster provisioner which is taking care
                                            of provisioning environments with generic-worker
                                            running on them. [default: test-provisioner]
          purgeCacheBaseURL                 The base URL for purge cache API calls.
                                            If not provided, the base URL for API calls is
                                            instead derived from rootURL setting as follows:
                                              * https://purge-cache.taskcluster.net/v1 for rootURL https://taskcluster.net
                                              * <rootURL>/api/purge-cache/v1 for all other rootURLs
          queueBaseURL                      The base URL for API calls to the queue service.
                                            If not provided, the base URL for API calls is
                                            instead derived from rootURL setting as follows:
                                              * https://queue.taskcluster.net/v1 for rootURL https://taskcluster.net
                                              * <rootURL>/api/queue/v1 for all other rootURLs
          region                            The EC2 region of the worker. Used by chain of trust.
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
                                            the current OS user will be used. [default: ` + strconv.FormatBool(runtime.GOOS != "windows") + `]
          secretsBaseURL                    The base URL for taskcluster secrets API calls.
                                            If not provided, the base URL for API calls is
                                            instead derived from rootURL setting as follows:
                                              * https://secrets.taskcluster.net/v1 for rootURL https://taskcluster.net
                                              * <rootURL>/api/secrets/v1 for all other rootURLs
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
                                            created on the worker. [default: ` + defaultTasksDir() + `]
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
          wstAudience                       The audience value for which to request websocktunnel
                                            credentials, identifying a set of WST servers this
                                            worker could connect to.  Optional if not using websocktunnel
                                            to expose live logs.
          wstServerURL                      The URL of the websocktunnel server with which to expose
                                            live logs.  Optional if not using websocktunnel to expose
                                            live logs.

        If an optional config setting is not provided in the json configuration file, the
        default will be taken (defaults documented above).

        If no value can be determined for a required config setting, the generic-worker will
        exit with a failure message.

  Exit Codes:

    0      Tasks completed successfully; no more tasks to run (see config setting
           numberOfTasksToRun).
    64     Not able to load generic-worker config. This could be a problem reading the
           generic-worker config file on the filesystem, a problem talking to AWS/GCP
           metadata service, or a problem retrieving config/files from the taskcluster
           secrets service.
    65     Not able to install generic-worker on the system.
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
    75     Not able to create an ed25519 key pair.
    76     Not able to save generic-worker config file after fetching it from AWS provisioner
           or Google Cloud metadata.
    77     Not able to apply required file access permissions to the generic-worker config
           file so that task users can't read from or write to it.
`
}
