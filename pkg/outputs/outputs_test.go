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
package outputs_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/outputs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestJSONFilter(t *testing.T) {

	testCases := []struct {
		name           string
		input          string
		spec           string
		result         string
		wantApplyError bool
	}{
		{
			name:  "Simple Valid Path",
			input: `{"foo":{"bar":"baz"}}`,
			spec: `filters:
  - json_path: foo.bar`,
			result:         "baz",
			wantApplyError: false,
		},
		{
			name:  "Valid Path But Not Found",
			input: `{"foo":{"bar":"baz"}}`,
			spec: `filters:
  - json_path: a.b`,
			wantApplyError: true,
		},
		{
			name:  "Invalid Path",
			input: `{"foo":{"bar":"baz"}}`,
			spec: `filters:
  - json_path: a.....b`,
			wantApplyError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var spec outputs.Spec
			err := yaml.Unmarshal([]byte(tc.spec), &spec)
			require.NoError(t, err)

			result, err := spec.Apply(tc.input)
			if tc.wantApplyError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.result, result)
		})
	}
}

func TestParse(t *testing.T) {
	input := `{"foo":{"bar":"baz"},"a":"b"}`
	specs := map[string]outputs.Spec{
		"first": {
			Filters: []outputs.Filter{
				&outputs.JSONFilter{
					Path: "foo.bar",
				},
			},
		},
		"second": {
			Filters: []outputs.Filter{
				&outputs.JSONFilter{
					Path: "a",
				},
			},
		},
	}

	results, err := outputs.Parse(specs, input)
	require.NoError(t, err)
	require.Equal(t, 2, len(results), "should have two outputs")
	assert.Equal(t, "baz", results["first"], "first output should be correct")
	assert.Equal(t, "b", results["second"], "second output should be correct")
}
