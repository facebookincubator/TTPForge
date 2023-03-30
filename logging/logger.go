package logging

import (
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the log level.
type LogLevel int8

// Available log levels.
const (
	LogLevelInfo LogLevel = iota
	LogLevelError
)

// Logger provides logging throughout the TTP Forge.
var Logger *zap.Logger

// AtomLevel provides an atomically changeable, dynamic logging level.
var AtomLevel zap.AtomicLevel
var cfg zap.Config
var err error

func init() {
	// https://github.com/uber-go/zap/issues/648
	// https://github.com/uber-go/zap/pull/307
	AtomLevel = zap.NewAtomicLevel()
	cfg = zap.NewDevelopmentConfig()
	cfg.Level = AtomLevel
	// use sugared logger
	Logger, err = cfg.Build()
	if err != nil {
		panic("failed to build logger")
	}
}

// InitLog initializes the TTP Forge's log file.
func InitLog(nocolor bool, logfile string, verbose bool) error {
	if !nocolor {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
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
	// use sugared logger
	Logger, err = cfg.Build()
	if err != nil {
		return err
	}

	if verbose {
		AtomLevel.SetLevel(zap.DebugLevel)
	}
	return nil
}
