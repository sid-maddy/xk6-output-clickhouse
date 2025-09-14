// Package clickhouse registers the extension for output.
package clickhouse

import (
	"fmt"

	"go.k6.io/k6/output"

	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse"
)

//nolint:gochecknoinits // Required to register the extension with k6.
func init() {
	output.RegisterExtension("clickhouse", func(p output.Params) (output.Output, error) {
		ext, err := clickhouse.New(&p)
		if err != nil {
			return nil, fmt.Errorf("error initlalizing the ClickHouse output extension: %w", err)
		}

		return ext, nil
	})
}
