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

package args_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/args"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateArgs(t *testing.T) {

	testCases := []struct {
		name           string
		specs          []args.Spec
		argKvStrs      []string
		expectedResult map[string]string
		wantError      bool
	}{
		{
			name: "Parse String Argument",
			specs: []args.Spec{
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
			expectedResult: map[string]string{
				"alpha": "foo",
				"beta":  "3",
			},
			wantError: false,
		},
		{
			name: "Handle Extra Equals",
			specs: []args.Spec{
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
			expectedResult: map[string]string{
				"alpha": "foo",
				"beta":  "bar=baz",
			},
			wantError: false,
		},
		{
			name: "Invalid Inputs (no '=')",
			specs: []args.Spec{
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
			name: "Invalid Inputs (Empty Value)",
			specs: []args.Spec{
				{
					Name: "alpha",
				},
			},
			argKvStrs: []string{
				"alpha=",
			},
			wantError: true,
		},
		{
			name: "Argument Name Not In Specs",
			specs: []args.Spec{
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
			specs: []args.Spec{
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
			specs: []args.Spec{
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := args.ParseAndValidate(tc.specs, tc.argKvStrs)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, args)
		})
	}

}
