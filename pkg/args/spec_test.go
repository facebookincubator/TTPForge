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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validateTestCase struct {
	name           string
	specs          []Spec
	argKvStrs      []string
	expectedResult map[string]any
	wantError      bool
}

func checkValidateTestCase(t *testing.T, tc validateTestCase) {
	args, err := ParseAndValidate(tc.specs, tc.argKvStrs)
	if tc.wantError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, tc.expectedResult, args)
}

func TestValidateArgs(t *testing.T) {

	testCases := []validateTestCase{
		{
			name: "Parse String and Integer Arguments",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
					Type: "int",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"beta=3",
			},
			expectedResult: map[string]any{
				"alpha": "foo",
				"beta":  3,
			},
			wantError: false,
		},
		{
			name: "Parse String and Integer Argument (Default Value)",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name:    "beta",
					Type:    "int",
					Default: "1337",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
			},
			expectedResult: map[string]any{
				"alpha": "foo",
				"beta":  1337,
			},
			wantError: false,
		},
		{
			name: "Handle Extra Equals",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"beta=bar=baz",
			},
			expectedResult: map[string]any{
				"alpha": "foo",
				"beta":  "bar=baz",
			},
			wantError: false,
		},
		{
			name: "Invalid Inputs (no '=')",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"wut",
			},
			wantError: true,
		},
		{
			name: "Invalid Inputs (Missing Required Argument)",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
			},
			wantError: true,
		},
		{
			name: "Argument Name Not In Specs",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"gamma=bar",
			},
			wantError: true,
		},
		{
			name: "Duplicate Name in Specs",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "alpha",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
			},
			wantError: true,
		},
		{
			name: "Wrong Type (string instead of int)",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
					Type: "int",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"beta=bar",
			},
			wantError: true,
		},
		{
			name: "Default Value Wrong Type (string instead of int)",
			specs: []Spec{
				{
					Name:    "alpha",
					Type:    "int",
					Default: "wut",
				},
			},
			argKvStrs: []string{
				"alpha=1337",
			},
			wantError: true,
		},
		{
			name: "Format with valid value",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "[A-Z_]+",
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT",
			},
			expectedResult: map[string]any{
				"alpha": "CECI_NEST_PAS_UNE_INT",
			},
			wantError: false,
		},
		{
			name: "Format with invalid value",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "[A-Z_]+",
				},
			},
			argKvStrs: []string{
				"alpha=INVALID_CECI_NEST_PAS-",
			},
			wantError: true,
		},
		{
			name: "Format with proper end and beginning tags",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "^[A-Z_-]+$",
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT-",
			},
			expectedResult: map[string]any{
				"alpha": "CECI_NEST_PAS_UNE_INT-",
			},
			wantError: false,
		},
		{
			name: "Format with improper regex",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "^[A-Z_-]+[$$",
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT-",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkValidateTestCase(t, tc)
		})
	}

}
