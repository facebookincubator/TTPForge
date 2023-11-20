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

package platforms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompatibleWith(t *testing.T) {

	testCases := []struct {
		name                string
		spec                Spec
		otherSpec           Spec
		expectValidateError bool
		desiredResult       bool
		correctString       string
	}{
		{
			name: "Empty OS and Architecture",
			spec: Spec{},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			expectValidateError: true,
		},
		{
			name: "Invalid OS",
			spec: Spec{
				OS: "invalid",
			},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			expectValidateError: true,
		},
		{
			name: "Invalid Arch",
			spec: Spec{
				Arch: "invalid",
			},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			expectValidateError: true,
		},
		{
			name: "Matching OS",
			spec: Spec{
				OS: "linux",
			},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			desiredResult: true,
			correctString: "linux/[any architecture]",
		},
		{
			name: "Matching OS but Not Architecture",
			spec: Spec{
				OS:   "linux",
				Arch: "arm64",
			},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			desiredResult: false,
			correctString: "linux/arm64",
		},
		{
			name: "Matching Architecture but not OS",
			spec: Spec{
				OS:   "darwin",
				Arch: "amd64",
			},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			desiredResult: false,
			correctString: "darwin/amd64",
		},
		{
			name: "Matching Architecture",
			spec: Spec{
				Arch: "amd64",
			},
			otherSpec: Spec{
				OS:   "linux",
				Arch: "amd64",
			},
			desiredResult: true,
			correctString: "[any OS]/amd64",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.expectValidateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			result := tc.spec.IsCompatibleWith(tc.otherSpec)
			assert.Equal(t, tc.desiredResult, result)
			assert.Equal(t, tc.correctString, tc.spec.String())
		})
	}
}
