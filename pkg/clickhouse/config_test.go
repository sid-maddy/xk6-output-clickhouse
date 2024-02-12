package clickhouse_test

import (
	"encoding/json"
	"testing"
	"time"

	clickhouse "github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/output"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	// TODO: add more cases
	testCases := map[string]struct {
		jsonRaw json.RawMessage
		env     map[string]string
		arg     string
		config  clickhouse.Config
		err     string
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
				"K6_CLICKHOUSE_ORG_ID":        "00000000-0000-0000-0000-000000000000",
				"K6_CLICKHOUSE_REGION":        "IOWA",
				"K6_CLICKHOUSE_RUN_ID":        "00000000-0000-0000-0000-000000000001",
				"K6_CLICKHOUSE_PUSH_INTERVAL": "4ms",
				"K6_CLICKHOUSE_LOG_LEVEL":     "debug",
			},
			config: clickhouse.Config{
				DSN:   "clickhouse://user:pass@localhost:9000/k6",
				Table: "k6_run_output",

				OrgID:  "00000000-0000-0000-0000-000000000000",
				Region: "IOWA",
				RunID:  "00000000-0000-0000-0000-000000000001",

				PushInterval: types.NullDurationFrom(4 * time.Millisecond),
				LogLevel:     logrus.DebugLevel,
			},
		},

		"early error": {
			env: map[string]string{
				"K6_CLICKHOUSE_DSN":           "clickhouse://user:pass@localhost:9000/k6",
				"K6_CLICKHOUSE_TABLE":         "k6_run_output",
				"K6_CLICKHOUSE_ORG_ID":        "00000000-0000-0000-0000-000000000000",
				"K6_CLICKHOUSE_REGION":        "IOWA",
				"K6_CLICKHOUSE_RUN_ID":        "abc",
				"K6_CLICKHOUSE_PUSH_INTERVAL": "4something",
				"K6_CLICKHOUSE_LOG_LEVEL":     "debug",
			},
			config: clickhouse.Config{
				DSN:   "clickhouse://user:pass@localhost:9000/k6",
				Table: "k6_run_output",

				OrgID:  "00000000-0000-0000-0000-000000000000",
				Region: "IOWA",
				RunID:  "00000000-0000-0000-0000-000000000001",

				PushInterval: types.NullDurationFrom(4 * time.Second),
				LogLevel:     logrus.DebugLevel,
			},
			err: `time: unknown unit "something" in duration "4something"`,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
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

			require.Equal(t, testCase.config.OrgID, config.OrgID)
			require.Equal(t, testCase.config.Region, config.Region)
			require.Equal(t, testCase.config.RunID, config.RunID)

			require.Equal(t, testCase.config.PushInterval, config.PushInterval)
			require.Equal(t, testCase.config.LogLevel, config.LogLevel)
		})
	}
}
