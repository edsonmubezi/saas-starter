package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the global Zap logger
func InitLogger(level string, format string, sampling bool) error {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "ts"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Sampling configuration (reduce noisy logs in production)
	if sampling {
		config.Sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	} else {
		config.Sampling = nil
	}

	// Build logger
	var err error
	Logger, err = config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	// Replace global logger
	zap.ReplaceGlobals(Logger)

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// WithFields returns a logger with additional fields
func WithFields(fields ...zap.Field) *zap.Logger {
	if Logger == nil {
		// Fallback to stderr if not initialized
		logger, _ := zap.NewProduction()
		return logger.With(fields...)
	}
	return Logger.With(fields...)
}

// InitDevelopmentLogger creates a development mode logger
func InitDevelopmentLogger() error {
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(Logger)
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Return a no-op logger if not initialized
		return zap.NewNop()
	}
	return Logger
}
