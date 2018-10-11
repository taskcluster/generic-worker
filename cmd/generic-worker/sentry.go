package main

import (
	"log"
	"runtime"
	"strconv"

	raven "github.com/getsentry/raven-go"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/tcauth"
)

func ReportCrashToSentry(r interface{}) {
	if config.SentryProject == "" {
		log.Println("No sentry project defined, not reporting to sentry")
		return
	}
	Auth := tcauth.New(
		&tcclient.Credentials{
			ClientID:    config.ClientID,
			AccessToken: config.AccessToken,
			Certificate: config.Certificate,
		},
	)
	Auth.BaseURL = config.AuthBaseURL
	res, err := Auth.SentryDSN(config.SentryProject)
	if err != nil {
		log.Printf("WARNING: Could not get sentry DSN: %v", err)
		return
	}
	client, err := raven.New(res.Dsn.Secret)
	if err != nil {
		log.Printf("Could not create raven client for reporting to sentry: %v", err)
		return
	}
	_, _ = client.CapturePanicAndWait(
		func() {
			panic(r)
		},
		map[string]string{
			"cleanUpTaskDirs":       strconv.FormatBool(config.CleanUpTaskDirs),
			"deploymentId":          config.DeploymentID,
			"instanceType":          config.InstanceType,
			"runTasksAsCurrentUser": strconv.FormatBool(config.RunTasksAsCurrentUser),
			"workerGroup":           config.WorkerGroup,
			"workerId":              config.WorkerID,
			"workerType":            config.WorkerType,
			"gwVersion":             version,
			"gwRevision":            revision,
			"GOOS":                  runtime.GOOS,
			"GOARCH":                runtime.GOARCH,
		},
	)
}
