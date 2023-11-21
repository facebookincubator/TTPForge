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

package args

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateArgsPath(t *testing.T) {

	// used in some test cases
	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	// cd into this directory and use it to verify
	// that relative paths are expanded appropriately
	tmpDir, err := os.MkdirTemp("", "ttpforge-testing")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	wd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		// this will break a bunch of other tests so be loud about it
		if err := os.Chdir(wd); err != nil {
			panic(err)
		}
	}()
	// macOS /private/var/...  shenanigans
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	testCases := []validateTestCase{
		{
			name: "Absolute Path",
			specs: []Spec{
				{
					Name: "mypath",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"mypath=/tmp/foo",
			},
			expectedResult: map[string]any{
				"mypath": "/tmp/foo",
			},
		},
		{
			name: "Path Containing Tilde",
			specs: []Spec{
				{
					Name: "mypath",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"mypath=~/wut",
			},
			expectedResult: map[string]any{
				"mypath": filepath.Join(homedir, "wut"),
			},
		},
		{
			name: "Relative Path",
			specs: []Spec{
				{
					Name: "mypath",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"mypath=" + filepath.Join("foo", "bar"),
			},
			// can't do a strict equality 0
			expectedResult: map[string]any{
				"mypath": filepath.Join(tmpDir, "foo", "bar"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkValidateTestCase(t, tc)
		})
	}
}
