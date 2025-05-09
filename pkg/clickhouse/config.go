package clickhouse

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/output"
)

// Config is the config for the ClickHouse output extension.
type Config struct {
	DSN   string `json:"dsn"`
	RunID string `json:"run_id"`

	PushInterval  types.NullDuration `json:"push_interval"`
	LogLevel      logrus.Level       `json:"log_level"`
	ClientOptions *clickhouse.Options
}

// NewConfig creates a new Config instance from the provided output.Params.
func NewConfig(params output.Params) (*Config, error) {
	cfg := Config{
		PushInterval: types.NullDurationFrom(1 * time.Second),
	}

	// Apply from JSON
	rawJSONConfig := params.JSONConfig
	if rawJSONConfig != nil {
		var jsonConfig Config
		if err := json.Unmarshal(rawJSONConfig, &jsonConfig); err != nil {
			return nil, fmt.Errorf("could not unmarshal JSON config: %w", err)
		}

		cfg.apply(jsonConfig)
	}

	// Apply from environment
	rawEnvConfig := params.Environment
	if len(rawEnvConfig) > 0 {
		for k, v := range rawEnvConfig {
			switch k {
			case "K6_CLICKHOUSE_DSN":
				cfg.DSN = v

			case "K6_CLICKHOUSE_RUN_ID":
				cfg.RunID = v

			case "K6_CLICKHOUSE_PUSH_INTERVAL":
				pushInterval, err := time.ParseDuration(v)
				if err != nil {
					return nil, fmt.Errorf("could not parse environment variable 'K6_CLICKHOUSE_PUSH_INTERVAL': %w", err)
				}

				cfg.PushInterval = types.NullDurationFrom(pushInterval)

			case "K6_CLICKHOUSE_LOG_LEVEL":
				var err error

				cfg.LogLevel, err = logrus.ParseLevel(value)
				if err != nil {
					return nil, fmt.Errorf("could not parse environment variable 'K6_CLICKHOUSE_LOG_LEVEL': %w", err)
				}
			}
		}
	}

	// Apply from CLI argument
	rawArg := params.ConfigArgument
	if rawArg != "" {
		cfg.DSN = rawArg
	}

	// Derive client config from address
	if cfg.DSN != "" {
		var err error

		cfg.ClientOptions, err = parseDSN(cfg.DSN)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

// apply merges the given configuration with the current configuration.
func (c Config) apply(cfg Config) Config {
	if cfg.DSN != "" {
		c.DSN = cfg.DSN
	}

	if cfg.RunID != "" {
		c.RunID = cfg.RunID
	}

	if cfg.PushInterval.Valid {
		c.PushInterval = cfg.PushInterval
	}

	return c
}

// parseDSN derives ClickHouse client options from the given DSN.
func parseDSN(dsn string) (*clickhouse.Options, error) {
	clientOptions, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not parse DSN: %w", err)
	}

	return clientOptions, nil
}
