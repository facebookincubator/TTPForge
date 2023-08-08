/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

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
