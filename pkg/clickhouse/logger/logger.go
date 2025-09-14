// Package logger provides the logger.
package logger

import (
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse/config"
)

var errInitializingLogger = errors.New("error initializing logger")

// New creates a new logger instance.
func New(cfg *config.Config, out io.Writer) (*zap.Logger, error) {
	loggerCfg := zap.NewProductionConfig()

	loggerCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerCfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	loggerCfg.Level = zap.NewAtomicLevelAt(cfg.LogLevel)

	logger, err := loggerCfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		//nolint:errcheck // The error returned in the next line is fatal anyway.
		fmt.Fprintln(out, errInitializingLogger.Error())

		return nil, fmt.Errorf("%w: %w", errInitializingLogger, err)
	}

	zap.RedirectStdLog(logger)

	logger.Info("Logger initialized", zap.String("level", cfg.LogLevel.String()))

	return logger, nil
}
