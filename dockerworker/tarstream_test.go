// +build docker

package dockerworker

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRewriteTarStream(t *testing.T) {
	// Create a big random blob
	blob := make([]byte, 6*1024*1024)
	_, err := rand.Read(blob)
	require.NoError(t, err)

	// Create tar file
	var input = []struct {
		Name string
		Data []byte
	}{
		{"hello.txt", []byte("hello-world")},
		{"rewritten.txt", []byte("data-to-be-rewritten")},
		{"filtered.txt", blob},
		{"to-rename.txt", []byte("data-in-renamed-file")},
		{"blob.bin", blob},
	}
	b := bytes.NewBuffer(nil)
	tw := tar.NewWriter(b)
	for _, f := range input {
		err = tw.WriteHeader(&tar.Header{
			Name: f.Name,
			Mode: 0600,
			Size: int64(len(f.Data)),
		})
		require.NoError(t, err)
		_, err = tw.Write(f.Data)
		require.NoError(t, err)
	}

	// Rewrite the tar file as stream
	out := bytes.NewBuffer(nil)
	err = rewriteTarStream(b, out, func(hdr *tar.Header, r io.Reader) (*tar.Header, io.Reader, error) {
		switch hdr.Name {
		case "hello.txt":
			return hdr, r, nil
		case "rewritten.txt":
			data := []byte("data-that-was-rewritten")
			hdr.Size = int64(len(data))
			return hdr, bytes.NewReader(data), nil
		case "filtered.txt":
			return nil, nil, nil
		case "to-rename.txt":
			hdr.Name = "was-renamed.txt"
			return hdr, r, nil
		case "blob.bin":
			return hdr, r, nil
		default:
			return nil, nil, errors.New("unhandled file name in test case")
		}
	})
	require.NoError(t, err)

	// Read the rewritten tar-stream
	tr := tar.NewReader(out)

	// declare expected output
	var output = []struct {
		Name string
		Data string
	}{
		{"hello.txt", "hello-world"},
		{"rewritten.txt", "data-that-was-rewritten"},
		{"was-renamed.txt", "data-in-renamed-file"},
		{"blob.bin", string(blob)},
	}
	for _, f := range output {
		fmt.Printf(" - verify %s\n", f.Name)
		var hdr *tar.Header
		hdr, err = tr.Next()
		require.NoError(t, err)
		require.Equal(t, f.Name, hdr.Name)
		var data []byte
		data, err = ioutil.ReadAll(tr)
		require.NoError(t, err)
		require.Equal(t, f.Data, string(data))
	}
	fmt.Println(" - verify that we have EOF")
	_, err = tr.Next()
	require.Equal(t, io.EOF, err)
}
