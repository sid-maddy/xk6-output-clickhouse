package clickhouse_test

import (
	"encoding/json"
	"testing"
	"time"

	clickhouse "github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse"
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
				PushInterval: types.NullDurationFrom(1 * time.Second),
			},
		},

		"overwrite": {
			env: map[string]string{
				"K6_CLICKHOUSE_DSN":           "clickhouse://user:pass@localhost:9000/k6",
				"K6_CLICKHOUSE_RUN_ID":        "abc",
				"K6_CLICKHOUSE_PUSH_INTERVAL": "4ms",
			},
			config: clickhouse.Config{
				DSN:          "clickhouse://user:pass@localhost:9000/k6",
				RunID:        "abc",
				PushInterval: types.NullDurationFrom(4 * time.Millisecond),
			},
		},

		"early error": {
			env: map[string]string{
				"K6_CLICKHOUSE_DSN":           "clickhouse://user:pass@localhost:9000/k6",
				"K6_CLICKHOUSE_RUN_ID":        "abc",
				"K6_CLICKHOUSE_PUSH_INTERVAL": "4something",
			},
			config: clickhouse.Config{
				DSN:          "clickhouse://user:pass@localhost:9000/k6",
				RunID:        "abc",
				PushInterval: types.NullDurationFrom(4 * time.Second),
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
			require.Equal(t, testCase.config.PushInterval, config.PushInterval)
		})
	}
}
