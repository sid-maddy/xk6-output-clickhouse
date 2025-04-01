// Package clickhouse registers the extension for output
package clickhouse

import (
	"fmt"

	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse"
	"go.k6.io/k6/output"
)

func init() {
	output.RegisterExtension("clickhouse", func(p output.Params) (output.Output, error) {
		var err error

		o, err := clickhouse.New(p)
		if err != nil {
			return nil, fmt.Errorf("could not initialize ClickHouse extension: %w", err)
		}

		return o, nil
	})
}
