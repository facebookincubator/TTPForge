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

package cmd

import (
	"os"
	"path/filepath"

	"testing"

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCreateTTP(t *testing.T) {
	testCases := []struct {
		name            string
		description     string
		targetPath      string
		wantCreateError bool
	}{
		{
			name:        "basic",
			description: "verify that the TTP is created with the correct contents",
			targetPath:  "anotherone.yaml",
		},
		{
			name:            "file already exists",
			description:     "should fail if we try to create a TTP on top of an existing file",
			targetPath:      repos.RepoConfigFileName,
			wantCreateError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// make disposable temp dir for testing
			tmpDir, err := os.MkdirTemp("", "ttpforge-testing")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)
			repoConfigFilePath := filepath.Join(tmpDir, repos.RepoConfigFileName)
			fsys := afero.NewOsFs()
			f, err := fsys.Create(repoConfigFilePath)
			require.NoError(t, err)
			defer f.Close()
			f.Write([]byte("ttp_search_paths:\n  - ttps\n"))

			// create the TTP
			createCmd := BuildRootCommand(nil)
			newFilePath := filepath.Join(tmpDir, tc.targetPath)
			createCmd.SetArgs([]string{"create", "ttp", newFilePath})
			err = createCmd.Execute()
			if tc.wantCreateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// confirm that it is a valid YAML that we can run
			runCmd := BuildRootCommand(nil)
			runCmd.SetArgs([]string{"run", newFilePath})
			err = runCmd.Execute()
			require.NoError(t, err)
		})
	}
}
