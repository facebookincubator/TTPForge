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
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnumTTPs(t *testing.T) {
	testConfigFilePath := filepath.Join("test-resources", "test-config.yaml")
	testCases := []struct {
		name             string
		description      string
		platform         string
		repo             string
		tactic           string
		technique        string
		subTech          string
		verbose          bool
		fileNameToCreate string
		errorExpected    bool
	}{
		{
			name:          "Correct all inputs",
			description:   "Enum command with all valid inputs should not throw an error",
			platform:      "linux",
			repo:          "enum-repo",
			errorExpected: false,
		},
		{
			name:          "Correct all inputs with multiple platforms",
			description:   "Enum command with all valid inputs should not throw an error",
			platform:      "linux,darwin",
			repo:          "enum-repo",
			tactic:        "TA0005",
			verbose:       true,
			errorExpected: false,
		},
		{
			name:          "1 incorrect platform in multiple platforms",
			description:   "Enum command with all valid inputs should not throw an error",
			platform:      "linux,darwin,badtest",
			repo:          "enum-repo",
			tactic:        "TA0005",
			technique:     "T1555",
			errorExpected: true,
		},

		{
			name:          "Non-existent repo",
			description:   "Enum command with repo name not available as input should throw an error",
			platform:      "linux",
			repo:          "ABCD",
			tactic:        "TA0005",
			technique:     "T1555",
			subTech:       "T1555.001",
			errorExpected: true,
		},
		{
			name:          "Default options",
			description:   "Default options of any platform and examples as repo (repo doesn't exist) should throw an error",
			platform:      "",
			repo:          "",
			tactic:        "",
			technique:     "",
			subTech:       "",
			errorExpected: true,
		},
		{
			name:          "Invalid platform",
			description:   "Invalid platform should throw an error",
			platform:      "solaris",
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"enum", "ttps", "--platform", tc.platform, "--repo", tc.repo, "--tactic", tc.tactic, "--technique", tc.technique, "--sub-tech", tc.subTech}
			if tc.verbose {
				args = append(args, "--verbose")
			}
			args = append(args, "-c", testConfigFilePath)
			if tc.platform == "" && tc.repo == "" {
				args = []string{"enum", "ttps", "-c", testConfigFilePath}
			}
			rc := BuildRootCommand(nil)
			rc.SetArgs(args)
			fmt.Println(strings.Join(args, " "))
			err := rc.Execute()
			if tc.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
