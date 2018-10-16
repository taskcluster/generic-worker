// +build docker

package dockerworker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRewriteRepositories(t *testing.T) {
	data, err := rewriteRepositories("my-image", []byte(
		`{"busybox":{"latest":"374004614a75c2c4afd41a3050b5217e282155eb1eb7b4ce8f22aa9f4b17ee57"}}`,
	))
	require.NoError(t, err)
	require.Contains(t, string(data), "my-image")
	require.NotContains(t, string(data), "busybox")
}

func TestRewriteManifest(t *testing.T) {
	data, err := rewriteManifest("my-image", []byte(`[
		{
			"Config": "2b8fd9751c4c0f5dd266fcae00707e67a2545ef34f9a29354585f93dac906749.json",
			"RepoTags": ["busybox:latest"],
			"Layers": [
				"374004614a75c2c4afd41a3050b5217e282155eb1eb7b4ce8f22aa9f4b17ee57/layer.tar"
			]
		}
	]`))
	require.NoError(t, err)
	require.Contains(t, string(data), "my-image")
	require.NotContains(t, string(data), "busybox")
}
