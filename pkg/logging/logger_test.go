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

package logging

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitLog(t *testing.T) {
	t.Run("TestNoColor", func(t *testing.T) {
		core, recordedLogs := observer.New(zapcore.InfoLevel)
		Logger = zap.New(core)

		if err := InitLog(true, "", false, false); err != nil {
			t.Errorf("error running InitLog(): %v", err)
		}

		Logger = Logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core { return core }))

		Logger.Info("Test message")
		logs := recordedLogs.All()
		if len(logs) != 1 || logs[0].Message != "Test message" {
			t.Errorf("the Logger did not produce expected output: %+v", logs)
		}
	})

	t.Run("TestLogFile", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "logfile")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		defer os.Remove(tempFile.Name())

		if err := InitLog(true, tempFile.Name(), false, false); err != nil {
			t.Errorf("error running InitLog(): %v", err)
		}

		testMessage := "Test log message"
		Logger.Info(testMessage)

		content, err := os.ReadFile(tempFile.Name())
		if err != nil {
			t.Errorf("failed to read log file: %v", err)
		}

		if !strings.Contains(string(content), testMessage) {
			t.Errorf("log file does not contain expected message")
		}
	})

	t.Run("TestVerbose", func(t *testing.T) {
		core, recordedLogs := observer.New(zapcore.DebugLevel)
		Logger = zap.New(core)

		if err := InitLog(true, "", true, true); err != nil {
			t.Errorf("error running InitLog(): %v", err)
		}

		Logger = Logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core { return core }))

		Logger.Debug("Test debug message")
		logs := recordedLogs.All()
		if len(logs) != 1 || logs[0].Message != "Test debug message" {
			t.Errorf("the Logger did not produce expected debug output: %+v", logs)
		}
	})
}
