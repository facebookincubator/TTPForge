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

package fileutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	// Set a test environment variable
	os.Setenv("TEST_VAR", "testvalue")
	defer os.Unsetenv("TEST_VAR")

	testCases := []struct {
		name           string
		path           string
		expectedResult string
	}{
		{
			name:           "Path with tilde",
			path:           "~/foo",
			expectedResult: filepath.Join(homedir, "foo"),
		},
		{
			name:           "Path without tilde",
			path:           "/foo",
			expectedResult: "/foo",
		},
		{
			name:           "Path without tilde 2",
			path:           "foo",
			expectedResult: "foo",
		},
		{
			name:           "Path with environment variable",
			path:           "$TEST_VAR/foo",
			expectedResult: "testvalue/foo",
		},
		{
			name:           "Path with ${VAR} syntax",
			path:           "${TEST_VAR}/foo",
			expectedResult: "testvalue/foo",
		},
		{
			name:           "Path with tilde and env var",
			path:           "~/foo/$TEST_VAR",
			expectedResult: filepath.Join(homedir, "foo/testvalue"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExpandPath(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
