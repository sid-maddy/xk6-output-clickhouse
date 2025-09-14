package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/output"
	"go.uber.org/zap/zapcore"

	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	// TODO: add more cases.
	//nolint:exhaustruct // Tests with defaults, which may be overridden.
	testCases := map[string]struct {
		env map[string]string

		arg string
		err string

		jsonRaw json.RawMessage
		config  config.Config
	}{
		"default": {
			config: config.Config{
				Table: config.DefaultTableName,

				PushInterval: types.NullDurationFrom(config.DefaultPushInterval),
				LogLevel:     config.DefaultLogLevel,
			},
		},

		"overwrite": {
			env: map[string]string{
				"K6_CLICKHOUSE_DSN":           "clickhouse://user:pass@localhost:9000/k6",
				"K6_CLICKHOUSE_TABLE":         "k6_run_output",
				"K6_CLICKHOUSE_ACCOUNT_ID":    "00000000-0000-0000-0000-000000000001",
				"K6_CLICKHOUSE_REGION":        "asia-south1",
				"K6_CLICKHOUSE_RUN_ID":        "00000000-0000-0000-0000-000000000002",
				"K6_CLICKHOUSE_PUSH_INTERVAL": "4ms",
				"K6_CLICKHOUSE_LOG_LEVEL":     "debug",
			},

			config: config.Config{
				DSN:   "clickhouse://user:pass@localhost:9000/k6",
				Table: "k6_run_output",

				AccountID: "00000000-0000-0000-0000-000000000001",
				Region:    "asia-south1",
				RunID:     "00000000-0000-0000-0000-000000000002",

				PushInterval: types.NullDurationFrom(4 * time.Millisecond),
				LogLevel:     zapcore.DebugLevel,
			},
		},

		"early error": {
			env: map[string]string{
				"K6_CLICKHOUSE_DSN":           "clickhouse://user:pass@localhost:9000/k6",
				"K6_CLICKHOUSE_TABLE":         "k6_run_output",
				"K6_CLICKHOUSE_ACCOUNT_ID":    "00000000-0000-0000-0000-000000000001",
				"K6_CLICKHOUSE_REGION":        "asia-south1",
				"K6_CLICKHOUSE_RUN_ID":        "abc",
				"K6_CLICKHOUSE_PUSH_INTERVAL": "4something",
				"K6_CLICKHOUSE_LOG_LEVEL":     "debug",
			},

			err: `time: unknown unit "something" in duration "4something"`,

			config: config.Config{
				DSN:   "clickhouse://user:pass@localhost:9000/k6",
				Table: "k6_run_output",

				AccountID: "00000000-0000-0000-0000-000000000001",
				Region:    "asia-south1",
				RunID:     "00000000-0000-0000-0000-000000000002",

				PushInterval: types.NullDurationFrom(4 * time.Second),
				LogLevel:     zapcore.DebugLevel,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			//nolint:exhaustruct // Params has defaults, which may be overridden.
			cfg, err := config.New(&output.Params{Environment: testCase.env})

			if testCase.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.err)

				return
			}

			require.NoError(t, err)

			require.Equal(t, testCase.config.DSN, cfg.DSN)
			require.Equal(t, testCase.config.Table, cfg.Table)

			require.Equal(t, testCase.config.AccountID, cfg.AccountID)
			require.Equal(t, testCase.config.Region, cfg.Region)
			require.Equal(t, testCase.config.RunID, cfg.RunID)

			require.Equal(t, testCase.config.PushInterval, cfg.PushInterval)
			require.Equal(t, testCase.config.LogLevel, cfg.LogLevel)
		})
	}
}
