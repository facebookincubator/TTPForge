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
)

func TestValidateArgsChoices(t *testing.T) {

	testCases := []validateTestCase{
		{
			name: "Choices: Argument Value is a Valid Choice",
			specs: []Spec{
				{
					Name:    "alpha",
					Choices: []string{"foo", "bar"},
				},
			},
			argKvStrs: []string{
				"alpha=bar",
			},
			expectedResult: map[string]any{
				"alpha": "bar",
			},
			wantError: false,
		},
		{
			name: "Choices: Default Value is Not a Valid Choice",
			specs: []Spec{
				{
					Name:    "alpha",
					Choices: []string{"foo", "bar"},
					Default: "baz",
				},
			},
			argKvStrs: []string{
				"alpha=bar",
			},
			wantError: true,
		},
		{
			name: "Choices: Argument Value is Not a Valid Choice",
			specs: []Spec{
				{
					Name:    "alpha",
					Choices: []string{"foo", "bar"},
				},
			},
			argKvStrs: []string{
				"alpha=baz",
			},
			wantError: true,
		},
		{
			name: "Choices: Testing Type Generality (int)",
			specs: []Spec{
				{
					Name:    "alpha",
					Type:    "int",
					Choices: []string{"1", "2"},
				},
			},
			argKvStrs: []string{
				"alpha=1",
			},
			expectedResult: map[string]any{
				"alpha": 1,
			},
			wantError: false,
		},
		{
			name: "Choices: Choice Configuration Contains Value with Incorrect Type",
			specs: []Spec{
				{
					Name:    "alpha",
					Type:    "int",
					Choices: []string{"CECI_NEST_PAS_UNE_INT", "1", "2"},
				},
			},
			argKvStrs: []string{
				"alpha=2",
			},
			wantError: true,
		},
		{
			name: "Choices: Wrong Type in Choice Configuration and Wrong Argument Type",
			specs: []Spec{
				{
					Name:    "alpha",
					Type:    "int",
					Choices: []string{"1", "2", "CECI_NEST_PAS_UNE_INT"},
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT",
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
