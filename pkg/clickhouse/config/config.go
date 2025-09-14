// Package config provides the configuration.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/output"
	"go.uber.org/zap/zapcore"
)

var errUnmarshallingJSONConfig = errors.New("error unmarshalling JSON config")

type errParsingEnvironmentVariableError struct {
	err  error
	name string
}

func (e errParsingEnvironmentVariableError) Error() string {
	return fmt.Sprintf("error parsing environment variable '%s': %s", e.name, e.err.Error())
}

func (e errParsingEnvironmentVariableError) Unwrap() error {
	return e.err
}

// Config is the extension's configuration.
type Config struct {
	ClientOptions *clickhouse.Options `json:"-"`

	DSN   string `json:"dsn"`
	Table string `json:"table"`

	AccountID string `json:"accountId"`
	RunID     string `json:"runId"`
	Region    string `json:"region"`

	PushInterval types.NullDuration `json:"pushInterval"`
	LogLevel     zapcore.Level      `json:"logLevel"`
}

// New creates a new [Config] instance.
func New(params *output.Params) (*Config, error) {
	//nolint:exhaustruct // Defaults, which may be overridden.
	cfg := Config{
		Table: DefaultTableName,

		PushInterval: types.NullDurationFrom(DefaultPushInterval),
		LogLevel:     DefaultLogLevel,
	}

	if err := cfg.applyFromJSON(params); err != nil {
		return nil, err
	}

	if err := cfg.applyFromEnv(params); err != nil {
		return nil, err
	}

	// Apply from CLI argument.
	rawArg := params.ConfigArgument
	if rawArg != "" {
		cfg.DSN = rawArg
	}

	// Derive client config from address.
	if cfg.DSN != "" {
		clientOptions, err := parseDSN(cfg.DSN)
		if err != nil {
			return nil, err
		}

		cfg.ClientOptions = clientOptions
	}

	return &cfg, nil
}

// apply merges the given configuration with the current configuration.
func (cfg *Config) apply(otherCfg *Config) *Config {
	if otherCfg.DSN != "" {
		cfg.DSN = otherCfg.DSN
	}

	if otherCfg.Table != "" {
		cfg.Table = otherCfg.Table
	}

	if otherCfg.AccountID != "" {
		cfg.AccountID = otherCfg.AccountID
	}

	if otherCfg.Region != "" {
		cfg.Region = otherCfg.Region
	}

	if otherCfg.RunID != "" {
		cfg.RunID = otherCfg.RunID
	}

	if otherCfg.PushInterval.Valid {
		cfg.PushInterval = otherCfg.PushInterval
	}

	return cfg
}

func (cfg *Config) applyFromJSON(params *output.Params) error {
	if params.JSONConfig != nil {
		var jsonConfig Config
		if err := json.Unmarshal(params.JSONConfig, &jsonConfig); err != nil {
			return fmt.Errorf("%w: %w", errUnmarshallingJSONConfig, err)
		}

		cfg.apply(&jsonConfig)
	}

	return nil
}

func (cfg *Config) applyFromEnv(params *output.Params) error {
	if len(params.Environment) > 0 {
		for key, value := range params.Environment {
			switch key {
			case "K6_CLICKHOUSE_DSN":
				cfg.DSN = value

			case "K6_CLICKHOUSE_TABLE":
				cfg.Table = value

			case "K6_CLICKHOUSE_ACCOUNT_ID":
				cfg.AccountID = value

			case "K6_CLICKHOUSE_RUN_ID":
				cfg.RunID = value

			case "K6_CLICKHOUSE_REGION":
				cfg.Region = value

			case "K6_CLICKHOUSE_PUSH_INTERVAL":
				pushInterval, err := time.ParseDuration(value)
				if err != nil {
					return errParsingEnvironmentVariableError{err: err, name: key}
				}

				cfg.PushInterval = types.NullDurationFrom(pushInterval)

			case "K6_CLICKHOUSE_LOG_LEVEL":
				logLevel, err := zapcore.ParseLevel(value)
				if err != nil {
					return errParsingEnvironmentVariableError{err: err, name: key}
				}

				cfg.LogLevel = logLevel
			}
		}
	}

	return nil
}

// parseDSN derives ClickHouse client options from the given DSN.
func parseDSN(dsn string) (*clickhouse.Options, error) {
	clientOptions, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	return clientOptions, nil
}
