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

package targets_test

import (
	"reflect"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/targets"
)

func TestParseAndValidateTargets(t *testing.T) {
	tests := []struct {
		name           string
		targetSpec     targets.TargetSpec
		expectedOutput map[string]interface{}
		expectedError  error
	}{
		{
			name: "OS and Arch only",
			targetSpec: targets.TargetSpec{
				OS:   []string{"linux", "windows"},
				Arch: []string{"x86", "amd64"},
			},
			expectedOutput: map[string]interface{}{
				"os":   []string{"linux", "windows"},
				"arch": []string{"x86", "amd64"},
			},
			expectedError: nil,
		},
		{
			name: "Cloud only",
			targetSpec: targets.TargetSpec{
				Cloud: []targets.Cloud{
					{
						Provider: "aws",
						Region:   "us-west-1",
					},
					{
						Provider: "gcp",
						Region:   "us-central1",
					},
				},
			},
			expectedOutput: map[string]interface{}{
				"cloud": []string{"aws:us-west-1", "gcp:us-central1"},
			},
			expectedError: nil,
		},
		{
			name: "Empty TargetSpec",
			targetSpec: targets.TargetSpec{
				OS:    nil,
				Arch:  nil,
				Cloud: nil,
			},
			expectedOutput: map[string]interface{}{},
			expectedError:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := targets.ParseAndValidateTargets(tc.targetSpec)

			// Check if the error is as expected
			if err != tc.expectedError {
				t.Fatalf("expected error %v, but got %v", tc.expectedError, err)
			}

			// Check if the output matches expected output
			if !reflect.DeepEqual(output, tc.expectedOutput) {
				t.Fatalf("expected output %v, but got %v", tc.expectedOutput, output)
			}
		})
	}
}
