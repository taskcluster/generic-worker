package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/taskcluster/httpbackoff"
	"github.com/taskcluster/taskcluster-client-go/tcsecrets"
)

type SecretFetcher struct {
	Secrets           *tcsecrets.Secrets
	RequestChannel    <-chan string
	DownloadDirectory string
	ProcessedChannel  chan<- string
}

type WorkerPool struct {
	Workers   []*SecretFetcher
	WaitGroup sync.WaitGroup
}

// This program is a simple command line utility. When you run it, it will
// spawn 20 go routines to download taskcluster secrets in parallel. It will
// put them in a sub-directory called secrets, creating it if necessary, with
// each file named as the secret. Path separators in the secret name will
// result in created subdirectories. I use this in combination with a cron job
// that runs
// https://github.com/petemoore/myscrapbook/blob/master/sync-secrets.sh every 5
// mins, in order to maintain a git history of secrets locally, in case
// something goes horribly wrong and we need to restore them.  I don't publish
// the git repository anywhere, obviously!
func main() {
	ss := tcsecrets.NewFromEnv()
	downloadDirectory := "secrets"
	requestChannel := make(chan string)
	processedChannel := make(chan string)
	_ = NewWorkerPool(20, requestChannel, processedChannel, ss, downloadDirectory)
	contToken := ""
	allSecrets := []string{}
	for {
		moreSecrets, err := ss.List(contToken, "")
		if err != nil {
			panic(err)
		}
		allSecrets = append(allSecrets, moreSecrets.Secrets...)
		contToken = moreSecrets.ContinuationToken
		if contToken == "" {
			break
		}
	}
	err := os.MkdirAll(downloadDirectory, 0755)
	if err != nil {
		panic(err)
	}
	go func() {
		defer close(requestChannel)
		for _, secret := range allSecrets {
			requestChannel <- secret
		}
	}()
	for completedSecret := range processedChannel {
		fmt.Println(completedSecret)
	}
}

func NewWorkerPool(capacity int, requestChannel <-chan string, processedChannel chan<- string, ss *tcsecrets.Secrets, downloadDirectory string) *WorkerPool {
	wp := &WorkerPool{}
	wp.WaitGroup.Add(capacity)
	wp.Workers = make([]*SecretFetcher, capacity, capacity)
	for i := 0; i < capacity; i++ {
		wp.Workers[i] = &SecretFetcher{
			Secrets:           ss,
			RequestChannel:    requestChannel,
			DownloadDirectory: downloadDirectory,
			ProcessedChannel:  processedChannel,
		}
		go func(i int) {
			wp.Workers[i].FetchUntilDone(&wp.WaitGroup)
		}(i)
	}
	go func() {
		wp.WaitGroup.Wait()
		close(processedChannel)
	}()
	return wp
}

func (secretFetcher *SecretFetcher) FetchUntilDone(wg *sync.WaitGroup) {
	for secret := range secretFetcher.RequestChannel {
		err := secretFetcher.fetch(secret)
		if err != nil {
			panic(err)
		}
	}
	wg.Done()
}

func (secretFetcher *SecretFetcher) fetch(secret string) (err error) {
	var u *url.URL
	u, err = secretFetcher.Secrets.Get_SignedURL(secret, time.Minute)
	if err != nil {
		return
	}
	var resp *http.Response
	resp, _, err = httpbackoff.Get(u.String())
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = resp.Body.Close()
		} else {
			resp.Body.Close()
		}
	}()
	var file *os.File
	path := filepath.Join(secretFetcher.DownloadDirectory, secret)
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	file, err = os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = file.Close()
		} else {
			file.Close()
		}
	}()
	io.Copy(file, resp.Body)
	secretFetcher.ProcessedChannel <- secret
	return
}
