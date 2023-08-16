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

package logging_test

import (
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitLog(t *testing.T) {
	t.Run("TestStacktrace", func(t *testing.T) {
		core, recordedLogs := observer.New(zapcore.InfoLevel)

		err := logging.InitLog(logging.Config{
			Stacktrace: true,
		})
		require.NoError(t, err)

		logger := logging.L().WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core { return core }))

		logger.Error("should produce a stack trace")

		entries := recordedLogs.All()
		require.Len(t, entries, 1)
		assert.Contains(t, entries[0].Stack, "logger_test.go", "stack trace should contain the test log file")
	})

	t.Run("TestLogFile", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "logfile")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		defer os.Remove(tempFile.Name())

		err = logging.InitLog(logging.Config{
			LogFile: tempFile.Name(),
		})
		require.NoError(t, err)

		testMessage := "Test log message"
		logging.L().Info(testMessage)

		content, err := os.ReadFile(tempFile.Name())
		require.NoError(t, err)
		assert.Contains(t, string(content), testMessage, "log file does not contain expected message")
	})

	t.Run("TestVerbose", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "logfile")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		defer os.Remove(tempFile.Name())

		err = logging.InitLog(logging.Config{
			LogFile: tempFile.Name(),
			Verbose: true,
		})
		require.NoError(t, err)

		testMessage := "debug: should show up for verbose"
		logging.L().Debug(testMessage)

		content, err := os.ReadFile(tempFile.Name())
		require.NoError(t, err)
		assert.Contains(t, string(content), testMessage, "log file does not contain expected message")
	})
}
