// +build docker

package dockerworker

import (
	"archive/tar"
	"io"

	"github.com/pkg/errors"
)

// a tarRewriter is a function that given a tar.Header and io.Reader returns a
// tar.Header and io.Reader to be written in place of the entry given.
//
// If a nil header is returned  the entry is skipped, if no transformation is
// desired the header and reader given can be returned with any changes.
type tarRewriter func(header *tar.Header, r io.Reader) (*tar.Header, io.Reader, error)

// rewriteTarStream rewrites a tar-stream read from r and written to w.
//
// For each entry in the tar-stream the rewriter function is called, if it
// returns an error then the process is aborted, if it returns a nil header the
// entry is skipped, otherwise the returned header and body is rewritten to the
// output tar-stream. Notice that rewriter can return the arguments given
// in-order to let entries pass-through.
func rewriteTarStream(r io.Reader, w io.Writer, rewriter tarRewriter) error {
	// Create a tar.Reader and tar.Writer, so we can move files between the two
	tr := tar.NewReader(r)
	tw := tar.NewWriter(w)
	for {
		// Read a tar.Header
		hdr, err := tr.Next()
		if err == io.EOF {
			break // we're done reading
		}
		if err != nil {
			return errors.Wrap(err, "failed to read tar.Reader while rewriting tar-stream")
		}

		// Allow the rewriter function to rewrite the entry
		var hdr2 *tar.Header
		var body io.Reader
		hdr2, body, err = rewriter(hdr, tr)
		if err != nil {
			return err
		}
		if hdr2 == nil {
			continue // skip this entry
		}

		// Write tar.Header and copy the body without making any changes
		err = tw.WriteHeader(hdr2)
		if err != nil {
			return errors.Wrap(err, "failed to write a tar.Header, while rewriting tar-stream")
		}
		// Copy file body to target as well
		_, err = io.Copy(tw, body)
		if err != nil {
			return errors.Wrap(err, "failed to write to tar.Writer, while rewriting tar-stream")
		}
	}
	// Close the tar.Writer, this ensure any cached bytes or outstanding errors are raised
	if err := tw.Close(); err != nil {
		return errors.Wrap(err, "failed to close tar.Writer")
	}
	return nil
}
