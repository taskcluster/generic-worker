// +build docker

package dockerworker

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/mattetti/filebuffer"
	"github.com/mitchellh/ioprogress"
	"github.com/taskcluster/slugid-go/slugid"
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

// ExtractArtifact gets files or directory trees from the container and copy them to destdir
func (d *DockerWorker) ExtractArtifact(container *docker.Container, path, destdir string) error {
	d.TaskLogger.Printf("Extracting '%s' from the container", path)

	tmp, err := ioutil.TempFile(os.TempDir(), slugid.Nice())
	if err != nil {
		return err
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	err = d.Client.DownloadFromContainer(container.ID, docker.DownloadFromContainerOptions{
		Context:      d.Context,
		Path:         path,
		OutputStream: tmp,
	})

	if err != nil {
		return fmt.Errorf("Error extracting '%s' from container: %v", path, err)
	}

	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	tr := tar.NewReader(tmp)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			d.TaskLogger.Print("Container extraction done")
			break
		}

		if err != nil {
			return err
		}

		d.TaskLogger.Printf("Unpacking '%s' at %s", hdr.Name, destdir)

		dest := filepath.Join(destdir, hdr.Name)

		if hdr.FileInfo().IsDir() {
			if err = os.MkdirAll(dest, 0700); err != nil && err != os.ErrExist {
				return err
			}
			continue
		}

		f, err := os.Create(dest)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, tr)

		// Why not use defer f.Close()?
		// Imagine we have to extract a dir with thousands of files,
		// if we use defer, the files will be closed only when we exit
		// the loop, which may, under a perfect storm, exaust the whole
		// memory available in the system.
		f.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
