package main

import (
	"io"
	"net/http"
	"sync"
	"time"
)

// WriteFlusher is a combination of io.Writer and http.Flusher, basically a
// stream that can be flushed.
type WriteFlusher interface {
	io.Writer
	http.Flusher
}

type nopFlusher struct {
	io.Writer
}

func (nopFlusher) Flush() {}

// NopFlusher returns a WriteFlusher implementation that wraps w, with Flush
// being a no-op.
func NopFlusher(w io.Writer) WriteFlusher {
	return nopFlusher{w}
}

type threadSafeWriteFlusher struct {
	m sync.Mutex
	w WriteFlusher
}

func (w *threadSafeWriteFlusher) Write(p []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.w.Write(p)
}

func (w *threadSafeWriteFlusher) Flush() {
	w.m.Lock()
	defer w.m.Unlock()
	w.w.Flush()
}

// CopyAndFlush will copy src to dst flushing at given interval.
func CopyAndFlush(dst WriteFlusher, src io.Reader, interval time.Duration) (int64, error) {
	w := &threadSafeWriteFlusher{w: dst}

	// Flush at every interval, until done is closed
	done := make(chan struct{})
	// Track that flushing has finished, before returning
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-time.After(interval):
				w.Flush()
			case <-done:
				wg.Done() // signal that flushing is done
				return
			}
		}
	}()

	// Copy from src to dst, through w
	n, err := io.Copy(w, src)

	// Stop flushing and wait for flushing thread to be done
	close(done)
	wg.Wait()

	// Do a final flush
	w.Flush()

	return n, err
}
