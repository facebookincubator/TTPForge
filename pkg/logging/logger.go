package logging

import (
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides logging throughout the TTP Forge.
var Logger *zap.Logger

// AtomLevel provides an atomically changeable, dynamic logging level.
var AtomLevel zap.AtomicLevel
var cfg zap.Config

func init() {
	// https://github.com/uber-go/zap/issues/648
	// https://github.com/uber-go/zap/pull/307
	AtomLevel = zap.NewAtomicLevel()
	cfg = zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.Level = AtomLevel
	// use sugared logger
	var err error
	Logger, err = cfg.Build()
	if err != nil {
		panic("failed to build logger")
	}
}

// InitLog initializes the TTP Forge's log file.
func InitLog(nocolor bool, logfile string, verbose bool, stacktrace bool) (err error) {
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

	if stacktrace {
		cfg.DisableStacktrace = false
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
		cfg.DisableStacktrace = true
	}
}