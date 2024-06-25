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

package logging

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestInitLogStacktraceNoLogfile verifies that the stack trace
// logging flag works correctly - it uses observer
// rather than a log file to test the "no log file" branch as well,
// which is why it is non-redundant with the file-based logging tests
// found later in this file
func TestInitLogStacktraceNoLogfile(t *testing.T) {
	core, recordedLogs := observer.New(zapcore.InfoLevel)

	initOnce = sync.Once{} // Reset the sync.Once for testing purposes
	err := InitLog(Config{
		Stacktrace: true,
	})
	require.NoError(t, err)

	logger := L().WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core { return core }))

	logger.Error("should produce a stack trace")

	entries := recordedLogs.All()
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Stack, "logger_test.go", "stack trace should contain the test log file")
}

// TestInitLog verifies that the various logging flags work -
// it uses a log file for all cases because
// flags like Verbose are easier to check when the output
// is sent to a file rather than through observer
func TestInitLog(t *testing.T) {

	tests := []struct {
		name      string
		config    Config
		logFunc   func(t *testing.T, testLogger *zap.SugaredLogger)
		checkFunc func(t *testing.T, logFileContents string)
	}{
		{
			name: "verbose",
			config: Config{
				Verbose: true,
			},
			logFunc: func(t *testing.T, testLogger *zap.SugaredLogger) {
				testLogger.Debug("hello, world")
			},
			checkFunc: func(t *testing.T, logFileContents string) {
				assert.Contains(t, logFileContents, "hello, world")
			},
		},
		{
			name: "no-color",
			config: Config{
				NoColor: true,
			},
			logFunc: func(t *testing.T, testLogger *zap.SugaredLogger) {
				testLogger.Info("no color")
			},
			checkFunc: func(t *testing.T, logFileContents string) {
				// ANSI reset code
				assert.NotContains(t, logFileContents, "\x1b[0m")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// use a temporary log file for each test case
			tempFile, err := os.CreateTemp("", "logger_test")
			require.NoError(t, err)
			logFile := tempFile.Name()
			defer os.Remove(logFile)
			cfg := tc.config
			cfg.LogFile = logFile

			initOnce = sync.Once{} // Reset the sync.Once for testing purposes
			err = InitLog(cfg)
			require.NoError(t, err)

			// actually create logs
			tc.logFunc(t, L())

			// read back result
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)
			tc.checkFunc(t, string(content))
		})
	}
}
