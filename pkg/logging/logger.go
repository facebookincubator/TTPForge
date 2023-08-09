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
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides logging throughout the TTPForge.
var Logger *zap.Logger

// AtomLevel provides an atomically changeable, dynamic logging level.
var AtomLevel zap.AtomicLevel
var cfg zap.Config

func init() {
	// https://github.com/uber-go/zap/issues/648
	// https://github.com/uber-go/zap/pull/307
	AtomLevel = zap.NewAtomicLevel()
	cfg = zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.Level = AtomLevel
	// use sugared logger
	var err error
	Logger, err = cfg.Build()
	if err != nil {
		panic("failed to build logger")
	}
}

// InitLog initializes the TTPForge's log file.
func InitLog(nocolor bool, logfile string, verbose bool) (err error) {
	if !nocolor {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// setup Logger to write to file if provided
	if logfile != "" {
		var fullpath string
		fullpath, err = filepath.Abs(logfile)
		if err != nil {
			return err
		}
		cfg.OutputPaths = append(cfg.OutputPaths, fullpath)
	}

	if verbose {
		AtomLevel.SetLevel(zap.DebugLevel)
	}

	// use sugared logger
	Logger, err = cfg.Build()
	if err != nil {
		return err
	}

	return nil
}

// ToggleDebug is used to trigger debug logs.
func ToggleDebug() {
	if AtomLevel.Level() != zap.DebugLevel {
		AtomLevel.SetLevel(zap.DebugLevel)
	} else {
		AtomLevel.SetLevel(zap.InfoLevel)
	}
}
