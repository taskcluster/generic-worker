package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/djherbis/stream"
	"github.com/taskcluster/taskcluster-base-go/scopes"
	tcclient "github.com/taskcluster/taskcluster-client-go"
)

type WebhookLogFeature struct{}

type WebhookLogTask struct {
	task           *TaskRun
	backingLogFile *os.File
	logStream      *stream.Stream
	detach         func()
}

func (feature *WebhookLogFeature) Name() string {
	return "Webhook Log"
}

func (feature *WebhookLogFeature) Initialise() error {
	return nil
}

func (feature *WebhookLogFeature) PersistState() error {
	return nil
}

func (feature *WebhookLogFeature) IsEnabled(fl EnabledFeatures) bool {
	return (TunnelServer != nil)
}

func (feature *WebhookLogFeature) NewTaskFeature(task *TaskRun) TaskFeature {
	return &WebhookLogTask{
		task: task,
	}
}

func (taskFeature *WebhookLogTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "X-Streaming")

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	wf, ok := w.(WriteFlusher)
	if ok {
		w.Header().Set("X-Streaming", "true")
	} else {
		wf = NopFlusher(w)
	}

	reader, err := taskFeature.logStream.NextReader()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	_, _ = CopyAndFlush(wf, reader, 100*time.Millisecond)
	_ = reader.Close()
}

func (taskFeature *WebhookLogTask) Start() *CommandExecutionError {
	logStream, err := stream.New("task_livelog_" + taskFeature.task.TaskID)
	if err != nil {
		return SetupFailed(err)
	}
	taskFeature.backingLogFile = taskFeature.task.logWriter.(*os.File)
	taskFeature.logStream = logStream

	_, seekErr := taskFeature.backingLogFile.Seek(0, 0)
	if seekErr != nil {
		log.Println("Could not seek to start of backing log file")
		return SetupFailed(seekErr)
	}

	_, copyErr := io.Copy(taskFeature.logStream, taskFeature.backingLogFile)
	if copyErr != nil {
		log.Println("Could not copy from backing log file")
		return SetupFailed(copyErr)
	}

	// the log writer now writes to both the log stream and backing log file
	taskFeature.task.logWriter = io.MultiWriter(taskFeature.backingLogFile, logStream)

	setCommandLogWriters(taskFeature.task.Commands, taskFeature.task.logWriter)

	uri, detach := TunnelServer.AttachHook(taskFeature)
	taskFeature.detach = detach
	// upload log artifact url
	// calculate log expiration deadline
	logExpirationDeadline := time.Time(taskFeature.task.TaskClaimResponse.Status.Runs[taskFeature.task.RunID].Started)
	logExpirationDeadline = logExpirationDeadline.Add(time.Duration(taskFeature.task.Payload.MaxRunTime) * time.Second)
	uploadErr := taskFeature.task.uploadArtifact(
		&RedirectArtifact{
			BaseArtifact: &BaseArtifact{
				Name:    livelogName,
				Expires: tcclient.Time(logExpirationDeadline),
			},
			MimeType: "text/plain; charset=utf-8",
			URL:      uri,
		},
	)
	if uploadErr != nil {
		log.Printf("error %v\n", uploadErr)
		return SetupFailed(uploadErr)
	}
	return nil
}

func (taskFeature *WebhookLogTask) Stop() *CommandExecutionError {
	// Close for good measure
	_ = taskFeature.logStream.Close()
	_ = taskFeature.logStream.Remove()

	// detach livelog hook
	if taskFeature.detach != nil {
		taskFeature.detach()
	}

	// redirect livelog to backing log. s3 artifact will be uploaded by taskRun
	uri := fmt.Sprintf("%v/task/%v/runs/%v/artifacts/%v", Queue.BaseURL, taskFeature.task.TaskID,
		taskFeature.task.RunID, livelogBackingName)
	err := taskFeature.task.uploadArtifact(
		&RedirectArtifact{
			BaseArtifact: &BaseArtifact{
				Name:    livelogName,
				Expires: taskFeature.task.Definition.Expires,
			},
			MimeType: "text/plain; charset=utf-8",
			URL:      uri,
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (taskFeature *WebhookLogTask) RequiredScopes() scopes.Required {
	return scopes.Required{}
}
