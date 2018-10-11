package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/taskcluster/httpbackoff"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
)

func main() {
	taskGroupID := os.Args[1]
	fmt.Printf("Task group: %v\n\n", taskGroupID)
	queue := tcqueue.NewFromEnv()
	query := func(ct string) *tcqueue.ListTaskGroupResponse {
		lgtr, err := queue.ListTaskGroup(taskGroupID, ct, "")
		if err != nil {
			log.Fatal(1, err)
		}
		for _, t := range lgtr.Tasks {
			show(queue, t)
		}
		return lgtr
	}
	ct := query("").ContinuationToken
	for ct != "" {
		ct = query(ct).ContinuationToken
	}
}

func show(queue *tcqueue.Queue, t tcqueue.TaskDefinitionAndStatus) {
	name := t.Task.ProvisionerID + "/" + t.Task.WorkerType + ":                                                                          "
	fmt.Print(name[:75])
	if t.Status.State == "pending" {
		fmt.Printf("not yet determined - task %v still pending...\n", t.Status.TaskID)
		return
	}
	var resp *http.Response
	artifactFound := ""
	for _, artifact := range []string{
		"public/logs/live_backing.log",
		"public/logs/chain_of_trust.log",
	} {
		logURL, err := queue.GetLatestArtifact_SignedURL(t.Status.TaskID, artifact, time.Hour)
		if err != nil {
			log.Fatal(2, err)
		}
		// log.Printf("URL: %v", logURL)
		resp, _, err = httpbackoff.Get(logURL.String())
		if err == nil {
			artifactFound = artifact
			break
		}
		switch e := err.(type) {
		case httpbackoff.BadHttpResponseCode:
			if e.HttpResponseCode == 404 {
				continue
			}
		}
		log.Fatal(3, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("*** ")
	}
	logContent := string(data)
	switch true {
	case strings.Contains(logContent, "Task not successful due to following exception"):
		fmt.Println("generic-worker - unknown version")
		// 	case strings.Contains(logContent, "Rejecting Schema: http://schemas.taskcluster.net/docker-worker/v1/payload.json"):
		//fmt.Println("worker: docker-worker - unknown version")
	case strings.Contains(logContent, "Worker Node Type:"):
		fmt.Println("docker-worker - unknown version")
	case strings.Contains(logContent, `"release": "https://github.com/taskcluster/generic-worker/releases/tag/v`):
		re := regexp.MustCompile(`"https://github.com/taskcluster/generic-worker/releases/tag/v([^"]*)"`)
		match := re.FindStringSubmatch(logContent)
		fmt.Printf("generic-worker %v\n", match[1])
	case strings.Contains(logContent, `not allowed at task.payload.features`):
		fmt.Println("taskcluster-worker - unknown version")
	case strings.Contains(logContent, `raise TaskVerificationError`):
		fmt.Println("scriptworker - unknown version")
	case strings.Contains(logContent, `KeyError: 'artifacts_deps'`):
		fmt.Println("some kind of scriptworker - unknown version")
	case artifactFound == "":
		fmt.Println("No artifacts found!")
	case artifactFound == "public/logs/chain_of_trust.log":
		fmt.Println("scriptworker chain of trust - unknown version")
	case strings.Contains(logContent, `os.environ.get('GITHUB_HEAD_REPO_URL', decision_json['payload']['env']['GITHUB_HEAD_REPO_URL'])`):
		fmt.Println("scriptworker - deepspeech - unknown version")
	default:
		fmt.Println("UNKNOWN")
		log.Fatal(5, logContent)
	}
}
