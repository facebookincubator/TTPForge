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

package blocks_test

import (
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
)

func TestLoadTTP(t *testing.T) {
	// Create a temporary file with valid TTP content.
	validTTPData := `
---
name: test_load_yaml
description: "Test loading and execution of YAML-based TTPs."
steps:
  - name: testing_exec
    file: ./tests/test.sh
    args:
      - blah
  - name: testing_exec_two
    file: ./tests/test.sh
    args:
      - blah
      - steps.testing_exec.output
`
	invalidTTPData := `
---
name: invalid_ttp
description: "Test detection and handling of invalid YAML-based TTPs."
steps:
  - name: "step1"
  - name: "step2"
`
	tmpFile1, err := os.CreateTemp(os.TempDir(), "valid_*.ttp")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := os.CreateTemp(os.TempDir(), "invalid_*.ttp")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile2.Name())

	if err := os.WriteFile(tmpFile1.Name(), []byte(validTTPData), 0644); err != nil {
		t.Fatalf("failed to write to %s: %v", tmpFile1.Name(), err)
	}

	if err := os.WriteFile(tmpFile2.Name(), []byte(invalidTTPData), 0644); err != nil {
		t.Fatalf("failed to write to %s: %v", tmpFile2.Name(), err)
	}

	testCases := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"Valid TTP", tmpFile1.Name(), false},
		{"Invalid TTP", tmpFile2.Name(), true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := blocks.LoadTTP(tc.filename); (err != nil) != tc.wantErr {
				t.Errorf("error running LoadTTP() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
