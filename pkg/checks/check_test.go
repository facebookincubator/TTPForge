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

package checks

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCheckVerify(t *testing.T) {

	testCases := []struct {
		name                 string
		contentStr           string
		fsysContents         map[string][]byte
		stepOutput           string
		expectUnmarshalError bool
		expectVerifyError    bool
	}{
		{
			name: "Check if Regular File Exists (Yes)",
			contentStr: `msg: File does not exist,
path_exists: should-exist.txt`,
			fsysContents: map[string][]byte{"should-exist.txt": []byte("foo")},
		},
		{
			name: "Check if Regular File Exists (No)",
			contentStr: `msg: File does not exist,
path_exists: does-not-exist.txt`,
			fsysContents:      map[string][]byte{"should-exist.txt": []byte("foo")},
			expectVerifyError: true,
		},
		{
			name: "path_exists + Checksum Verification (Success)",
			contentStr: `msg: File does not exists or does not have expected content,
path_exists: should-exist.txt
checksum:
  sha256: 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae`,
			fsysContents:         map[string][]byte{"should-exist.txt": []byte("foo")},
			expectUnmarshalError: false,
			expectVerifyError:    false,
		},
		{
			name: "Check if Checksum is Correct (Yes)",
			contentStr: `msg: File does not exists or does not have expected content,
path_exists: has-correct-hash.txt
checksum:
  sha256: 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae`,
			fsysContents: map[string][]byte{"has-correct-hash.txt": []byte("foo")},
		},
		{
			name: "Check if Checksum is Correct (Yes)",
			contentStr: `msg: File does not exists or does not have expected content,
path_exists: incorrect-hash.txt
checksum:
  sha256: "absolutely wrong"`,
			fsysContents:      map[string][]byte{"incorrect-hash.txt": []byte("foo")},
			expectVerifyError: true,
		},
		{
			name: "Content Contains (Success)",
			contentStr: `msg: File should contain expected string
path_exists: config.txt
content_contains: "enabled=true"`,
			fsysContents: map[string][]byte{"config.txt": []byte("setting1=value1\nenabled=true\nsetting2=value2")},
		},
		{
			name: "Content Contains (Failure)",
			contentStr: `msg: File should contain expected string
path_exists: config.txt
content_contains: "enabled=false"`,
			fsysContents:      map[string][]byte{"config.txt": []byte("setting1=value1\nenabled=true\nsetting2=value2")},
			expectVerifyError: true,
		},
		{
			name: "Content Not Contains (Success)",
			contentStr: `msg: File should not contain debug flags
path_exists: config.txt
content_not_contains: "DEBUG=1"`,
			fsysContents: map[string][]byte{"config.txt": []byte("setting1=value1\nenabled=true\nsetting2=value2")},
		},
		{
			name: "Content Not Contains (Failure)",
			contentStr: `msg: File should not contain debug flags
path_exists: config.txt
content_not_contains: "enabled"`,
			fsysContents:      map[string][]byte{"config.txt": []byte("setting1=value1\nenabled=true\nsetting2=value2")},
			expectVerifyError: true,
		},
		{
			name: "Content Regex (Success)",
			contentStr: `msg: File should match version pattern
path_exists: version.txt
content_regex: "v[0-9]+\\.[0-9]+\\.[0-9]+"`,
			fsysContents: map[string][]byte{"version.txt": []byte("version: v1.2.3")},
		},
		{
			name: "Content Regex (Failure)",
			contentStr: `msg: File should match version pattern
path_exists: version.txt
content_regex: "v[0-9]+\\.[0-9]+\\.[0-9]+"`,
			fsysContents:      map[string][]byte{"version.txt": []byte("version: v1.2.x")},
			expectVerifyError: true,
		},
		{
			name: "Content Regex Invalid Pattern",
			contentStr: `msg: Invalid regex should fail
path_exists: test.txt
content_regex: "[invalid"`,
			fsysContents:      map[string][]byte{"test.txt": []byte("test")},
			expectVerifyError: true,
		},
		{
			name: "Permissions Check (Success)",
			contentStr: `msg: File should have correct permissions
path_exists: script.sh
permissions: "0644"`,
			fsysContents: map[string][]byte{"script.sh": []byte("#!/bin/bash\necho 'hello'")},
		},
		{
			name: "Multiple Content Checks (Success)",
			contentStr: `msg: File should pass all content checks
path_exists: app.conf
content_contains: "enabled=true"
content_not_contains: "DEBUG"
content_regex: "enabled=[a-z]+"`,
			fsysContents: map[string][]byte{"app.conf": []byte("enabled=true\nport=8080")},
		},
		{
			name: "Checksum + Content Contains (Success)",
			contentStr: `msg: File should have correct hash and content
path_exists: payload.txt
checksum:
  sha256: 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae
content_contains: "foo"`,
			fsysContents: map[string][]byte{"payload.txt": []byte("foo")},
		},
		{
			name: "Output Contains (Pass)",
			contentStr: `msg: Running as root
output_contains: "root"`,
			stepOutput: "root\nweb-01\n",
		},
		{
			name: "Output Contains (Fail)",
			contentStr: `msg: Running as root
output_contains: "root"`,
			stepOutput:        "deploy\nweb-01\n",
			expectVerifyError: true,
		},
		{
			name: "Output Not Contains (Pass)",
			contentStr: `msg: No errors in output
output_not_contains: "error"`,
			stepOutput: "success\nall good\n",
		},
		{
			name: "Output Not Contains (Fail)",
			contentStr: `msg: No errors in output
output_not_contains: "error"`,
			stepOutput:        "something went error here\n",
			expectVerifyError: true,
		},
		{
			name: "Output Regex (Pass)",
			contentStr: `msg: Hostname matches pattern
output_regex: "web-[0-9]+"`,
			stepOutput: "root\nweb-01\n",
		},
		{
			name: "Output Regex (Fail)",
			contentStr: `msg: Hostname matches pattern
output_regex: "web-[0-9]+"`,
			stepOutput:        "root\ndb-server\n",
			expectVerifyError: true,
		},
		{
			name: "Output Regex Invalid Pattern",
			contentStr: `msg: Bad regex
output_regex: "[invalid"`,
			stepOutput:        "test output",
			expectVerifyError: true,
		},
		{
			name: "Combined Output Checks (Pass)",
			contentStr: `msg: Output meets all criteria
output_contains: "root"
output_not_contains: "error"
output_regex: "root"`,
			stepOutput: "root\nweb-01\n",
		},
		{
			name: "Output Contains with Empty Step Output (Fail)",
			contentStr: `msg: Should fail on empty output
output_contains: "something"`,
			stepOutput:        "",
			expectVerifyError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prep filesystem
			fsys, err := testutils.MakeAferoTestFs(tc.fsysContents)
			require.NoError(t, err)

			// decode the check
			var check Check
			err = yaml.Unmarshal([]byte(tc.contentStr), &check)
			if tc.expectUnmarshalError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// run verification
			err = check.Verify(VerificationContext{FileSystem: fsys, StepOutput: tc.stepOutput})
			if tc.expectVerifyError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

}
