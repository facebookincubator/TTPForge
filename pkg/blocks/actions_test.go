/*
Copyright © 2024-present, Meta Platforms, Inc. and affiliates
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

package blocks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTimeout(t *testing.T) {
	testCases := []struct {
		name          string
		stepTimeout   string
		longRunning   bool
		want          time.Duration
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name: "default when unset",
			want: DefaultExecutionTimeout,
		},
		{
			name:        "shorter than default needs no opt-in",
			stepTimeout: "30m",
			want:        30 * time.Minute,
		},
		{
			name:        "equal to default needs no opt-in",
			stepTimeout: "100m",
			want:        DefaultExecutionTimeout,
		},
		{
			name:        "longer than default with long_running",
			stepTimeout: "7h",
			longRunning: true,
			want:        7 * time.Hour,
		},
		{
			name:          "longer than default without long_running rejected",
			stepTimeout:   "7h",
			wantErr:       true,
			wantErrSubstr: "long_running",
		},
		{
			name:          "exceeds max even with long_running",
			stepTimeout:   "25h",
			longRunning:   true,
			wantErr:       true,
			wantErrSubstr: "maximum allowed",
		},
		{
			// Regression: when both the long_running gate and the max-cap
			// would reject the value, the user should see the max-cap error
			// first -- otherwise they'd be told to opt in via long_running
			// and then hit a second rejection on the next attempt.
			name:          "exceeds max without long_running reports max not opt-in",
			stepTimeout:   "25h",
			wantErr:       true,
			wantErrSubstr: "maximum allowed",
		},
		{
			name:          "zero rejected",
			stepTimeout:   "0s",
			wantErr:       true,
			wantErrSubstr: "must be > 0",
		},
		{
			name:          "negative rejected",
			stepTimeout:   "-5m",
			wantErr:       true,
			wantErrSubstr: "must be > 0",
		},
		{
			name:          "unparsable rejected",
			stepTimeout:   "not-a-duration",
			wantErr:       true,
			wantErrSubstr: "invalid step_timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td := timedActionDefaults{
				StepTimeout: tc.stepTimeout,
				LongRunning: tc.longRunning,
			}
			got, err := td.resolveTimeout()
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrSubstr != "" {
					assert.Contains(t, err.Error(), tc.wantErrSubstr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
