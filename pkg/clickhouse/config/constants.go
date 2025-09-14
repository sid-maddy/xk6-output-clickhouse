package config

import (
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	// DefaultTableName is the default output table name.
	DefaultTableName = "run_output"

	// DefaultPushInterval is the default metric push interval.
	DefaultPushInterval = 1 * time.Second

	// DefaultLogLevel is the default log level of the extension.
	DefaultLogLevel = zapcore.InfoLevel

	// DefaultConnectionVerificationTimeout is the default timeout for verifying the ClickHouse connection.
	DefaultConnectionVerificationTimeout = 10 * time.Second
)
