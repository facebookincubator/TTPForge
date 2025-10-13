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
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Config contains various formatting options for the global logger
type Config struct {
	Verbose    bool
	LogFile    string
	NoColor    bool
	Stacktrace bool
}

var (
	logger              *zap.SugaredLogger
	initOnce            sync.Once
	encoderRegisterOnce sync.Once
	indentLevel         int
)

// L returns the global logger for ttpforge
func L() *zap.SugaredLogger {
	// If the logger is not initialized at this point, initialize it with default settings.  Usually this should only happen in tests.
	if logger == nil {
		err := InitLog(Config{})
		if err != nil {
			// this should never fail - if it does
			// something weird happened so we panic
			panic(err)
		}
	}
	return logger
}

// DividerThick prints a divider line made of `=` characters
// to help with the readability of logs
func DividerThick() {
	L().Infow("========================================")
}

// DividerThin prints a divider line made of `-` characters
// to help with the readability of logs
func DividerThin() {
	L().Infow("----------------------------------------")
}

func IncreaseIndentLevel() {
	indentLevel++
}

func DecreaseIndentLevel() {
	if indentLevel > 0 {
		indentLevel--
	}
}

type indentedEncoder struct {
	zapcore.Encoder
	pool buffer.Pool
	cfg  zapcore.EncoderConfig
}

func (e *indentedEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := e.pool.Get()

	// Add indentations for subTTPs
	buf.AppendString(strings.Repeat("\t", indentLevel))

	consolebuf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(consolebuf.Bytes())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (e *indentedEncoder) Clone() zapcore.Encoder {
	// Needed to handle indents in structured logging
	return &indentedEncoder{
		Encoder: zapcore.NewConsoleEncoder(e.cfg),
		pool:    buffer.NewPool(),
		cfg:     e.cfg,
	}
}

// InitLog initializes TTPForge global logger
func InitLog(config Config) (err error) {
	// Register the indented encoder only once globally
	encoderRegisterOnce.Do(func() {
		err := zap.RegisterEncoder("indented", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
			enc := &indentedEncoder{
				Encoder: zapcore.NewConsoleEncoder(cfg),
				pool:    buffer.NewPool(),
				cfg:     cfg,
			}
			return enc, nil
		})
		if err != nil {
			panic(err) // Use panic here since sync.Once does not allow error return
		}
	})

	initOnce.Do(func() {
		zcfg := zap.NewDevelopmentConfig()
		if config.NoColor {
			zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		} else {
			zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		zcfg.Encoding = "indented"

		if config.LogFile != "" {
			fullpath, err := filepath.Abs(config.LogFile)
			if err != nil {
				panic(err) // Use panic here since sync.Once does not allow error return
			}
			zcfg.OutputPaths = append(zcfg.OutputPaths, fullpath)
		}

		if config.Verbose {
			zcfg.Level.SetLevel(zap.DebugLevel)
		} else {
			zcfg.Level.SetLevel(zap.InfoLevel)
			zcfg.EncoderConfig.CallerKey = zapcore.OmitKey
			zcfg.EncoderConfig.TimeKey = zapcore.OmitKey
		}
		if !config.Stacktrace {
			zcfg.DisableStacktrace = true
		}

		baseLogger, err := zcfg.Build()
		if err != nil {
			panic(err) // Use panic here since sync.Once does not allow error return
		}
		logger = baseLogger.Sugar()
	})
	return nil
}
