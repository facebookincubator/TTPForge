package repos_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigs(t *testing.T) {
	tests := []struct {
		name           string
		specs          []repos.Spec
		fileSystem     fs.FS
		correctConfigs []repos.Config
		expectError    bool
	}{
		{
			name: "two valid configs",
			specs: []repos.Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Name: "default",
					Path: "repos/d",
				},
			},
			fileSystem: fstest.MapFS{
				"repos/a/" + repos.RepoConfigFileName: &fstest.MapFile{
					Data: []byte(`ttp_search_paths: ["ttps_a"]`),
				},
				"repos/b/not-a-config": &fstest.MapFile{
					Data: []byte("foo"),
				},
				"repos/c/also-not-a-config": &fstest.MapFile{
					Data: []byte("bar"),
				},
				"repos/d/" + repos.RepoConfigFileName: &fstest.MapFile{
					Data: []byte(`ttp_search_paths: ["ttps_d"]`),
				},
			},
			correctConfigs: []repos.Config{
				{
					TTPSearchPaths: []string{"ttps_a"},
				},
				{
					TTPSearchPaths: []string{"ttps_d"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configs, err := repos.LoadConfigs(tc.fileSystem, tc.specs)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.correctConfigs, configs)
			}
		})
	}
}
