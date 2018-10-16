// +build docker

package dockerworker

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/mattetti/filebuffer"
	"github.com/mitchellh/ioprogress"
)

// DownloadArtifact downloads an artifact using exponential backoff algorithm
func (d *DockerWorker) DownloadArtifact(taskID, runID, name string, out io.WriteSeeker) (ret error) {
	var u *url.URL

	backoffError := backoff.Retry(func() (err error) {
		// rewind out stream
		_, ret = out.Seek(0, io.SeekStart)
		if ret != nil {
			return
		}

		if runID == "" {
			u, ret = d.Queue.GetLatestArtifact_SignedURL(taskID, name, 24*time.Hour)
		} else {
			u, ret = d.Queue.GetArtifact_SignedURL(taskID, runID, name, 24*time.Hour)
		}

		if ret != nil {
			return
		}

		d.TaskLogger.Printf("Downloading %s/%s/%s from %s\n", taskID, runID, name, u.String())

		// Build a custom request object so we can embed a context into it
		var req *http.Request
		req, ret = http.NewRequest(http.MethodGet, u.String(), nil)
		if ret != nil {
			return
		}
		req = req.WithContext(d.Context)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			buff := filebuffer.New([]byte{})

			// Build the error message with the status code and the response body
			errorMessage := fmt.Sprintf("Error downloading %s/%s, status=%s", taskID, name, resp.Status)
			if _, err2 := io.Copy(buff, resp.Body); err2 == nil {
				errorMessage += "\n" + buff.String()
			}

			// For status codes other than 5XX there is no point in issuing a retry
			if resp.StatusCode >= 300 && resp.StatusCode < 500 {
				ret = errors.New(errorMessage)
			} else {
				err = errors.New(errorMessage)
			}

			return
		}

		size, err2 := strconv.Atoi(resp.Header.Get("Content-Length"))
		if err2 != nil {
			ret = err2
			return
		}

		// Depending on the implementation, the body can also issue network requests,
		// that's why we need to retry if io.Copy fails
		_, err = io.Copy(out, &ioprogress.Reader{
			Reader:   resp.Body,
			Size:     int64(size),
			DrawFunc: ioprogress.DrawTerminal(d.LivelogWriter),
		})

		return
	}, backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), d.Context), 3))

	if ret == nil {
		ret = backoffError
	}

	return
}
