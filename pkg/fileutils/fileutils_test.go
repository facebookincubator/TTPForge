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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandTilde(t *testing.T) {

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	testCases := []struct {
		name           string
		path           string
		expectedResult string
	}{
		{
			name:           "No Tilde (Should Return Same Path)",
			path:           "/foo/bar",
			expectedResult: "/foo/bar",
		},
		{
			name:           "Leading Tilde (Should Be Expanded)",
			path:           "~/victory",
			expectedResult: homedir + "/victory",
		},
		{
			name:           "Trailing Tilde (Should Not Be Expanded)",
			path:           "do-not-expand/~",
			expectedResult: "do-not-expand/~",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExpandTilde(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
