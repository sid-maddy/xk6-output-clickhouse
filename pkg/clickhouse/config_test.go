package clickhouse_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/output"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	// TODO: add more cases
	testCases := map[string]struct {
		env map[string]string

		arg string
		err string

		jsonRaw json.RawMessage
		config  clickhouse.Config
	}{
		"default": {
			config: clickhouse.Config{
				Table: "run_output",

				PushInterval: types.NullDurationFrom(1 * time.Second),
				LogLevel:     logrus.InfoLevel,
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

			config: clickhouse.Config{
				DSN:   "clickhouse://user:pass@localhost:9000/k6",
				Table: "k6_run_output",

				AccountID: "00000000-0000-0000-0000-000000000001",
				Region:    "asia-south1",
				RunID:     "00000000-0000-0000-0000-000000000002",

				PushInterval: types.NullDurationFrom(4 * time.Millisecond),
				LogLevel:     logrus.DebugLevel,
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

			config: clickhouse.Config{
				DSN:   "clickhouse://user:pass@localhost:9000/k6",
				Table: "k6_run_output",

				AccountID: "00000000-0000-0000-0000-000000000001",
				Region:    "asia-south1",
				RunID:     "00000000-0000-0000-0000-000000000002",

				PushInterval: types.NullDurationFrom(4 * time.Second),
				LogLevel:     logrus.DebugLevel,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config, err := clickhouse.NewConfig(output.Params{Environment: testCase.env})

			if testCase.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.err)

				return
			}

			require.NoError(t, err)

			require.Equal(t, testCase.config.DSN, config.DSN)
			require.Equal(t, testCase.config.Table, config.Table)

			require.Equal(t, testCase.config.AccountID, config.AccountID)
			require.Equal(t, testCase.config.Region, config.Region)
			require.Equal(t, testCase.config.RunID, config.RunID)

			require.Equal(t, testCase.config.PushInterval, config.PushInterval)
			require.Equal(t, testCase.config.LogLevel, config.LogLevel)
		})
	}
}
