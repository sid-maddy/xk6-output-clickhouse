package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"
	"go.k6.io/k6/output"
)

const runOutputTableName = "run_output"

// Output implements the output.Output interface.
type Output struct {
	output.SampleBuffer

	config          Config
	conn            clickhouse.Conn
	periodicFlusher *output.PeriodicFlusher

	logger logrus.FieldLogger
}

var _ output.Output = new(Output)

// outputRow represents a row of the run output.
// NOTE: The fields are in snake_case to match the column names in the ClickHouse table.
type outputRow struct {
	RunID    string            `ch:"run_id"`
	Time     time.Time         `ch:"time"`
	Metric   string            `ch:"metric"`
	Value    float64           `ch:"value"`
	Tags     map[string]string `ch:"tags"`
	Metadata map[string]string `ch:"metadata"`
}

// New creates an instance of the emitter.
func New(p output.Params) (*Output, error) {
	logger := newLogger(p.StdOut)

	config, err := NewConfig(p)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	setLoggerLevel(logger, config.LogLevel)

	logger.Debug("opening connection to ClickHouse")

	conn, err := clickhouse.Open(config.ClientOptions)
	if err != nil {
		return nil, fmt.Errorf("could not connect to ClickHouse: %w", err)
	}

	return &Output{
		config: *config,
		conn:   conn,

		logger: logger,
	}, nil
}

// Description returns a human-readable description of the output that will be shown in `k6 run`.
func (o *Output) Description() string {
	return fmt.Sprintf("clickhouse (%s)", o.config.DSN)
}

// Start performs initialization tasks prior to the engine using the output.
func (o *Output) Start() error {
	o.logger.Debug("starting")
	defer o.logger.Debug("started")

	ctx := context.Background()

	database := o.config.ClientOptions.Auth.Database
	o.logger.WithField("db", database).Debug("verifying provided database")

	if err := o.conn.Exec(ctx, "USE "+database); err != nil {
		o.logger.WithField("db", database).WithError(err).Error("provided database does not exist")
		return fmt.Errorf("could not verify provided database: %w", err)
	}

	o.logger.Debug("verifying run output table")

	if err := o.conn.Exec(ctx, fmt.Sprintf("DESCRIBE TABLE %s.%s", database, o.config.Table)); err != nil {
		o.logger.
			WithFields(logrus.Fields{"db": database, "table": o.config.Table}).
			WithError(err).
			Error("run output table does not exist")

		return fmt.Errorf("could not verify run output table: %w", err)
	}

	pushInterval := o.config.PushInterval.TimeDuration()
	o.logger.WithField("push_interval", pushInterval).Debug("creating periodic flusher")

	var err error

	o.periodicFlusher, err = output.NewPeriodicFlusher(pushInterval, o.flushMetrics)
	if err != nil {
		return fmt.Errorf("could not create periodic flusher: %w", err)
	}

	return nil
}

// Stop flushes all remaining metrics and finalizes the test run.
func (o *Output) Stop() error {
	o.logger.Debug("stopping")
	defer o.logger.Debug("stopped")

	o.periodicFlusher.Stop()

	if err := o.conn.Close(); err != nil {
		return fmt.Errorf("could not close ClickHouse connection: %w", err)
	}

	return nil
}

// flushMetrics periodically flushes buffered metric samples to ClickHouse.
func (o *Output) flushMetrics() {
	ctx := context.Background()

	samples := o.GetBufferedSamples()
	if len(samples) == 0 {
		return
	}

	start := time.Now()

	o.logger.WithField("count", len(samples)).Debug("emitting samples")

	batch, err := o.conn.PrepareBatch(
		ctx,
		fmt.Sprintf("INSERT INTO %s.%s", o.config.ClientOptions.Auth.Database, runOutputTableName),
	)
	if err != nil {
		o.logger.WithError(err).Error("error preparing batch insert query")
		return
	}

	var count int

	for _, sc := range samples {
		samples := sc.GetSamples()
		count += len(samples)

		for _, sample := range samples {
			if err := batch.AppendStruct(&outputRow{
				RunID:    o.config.RunID,
				Time:     sample.Time,
				Metric:   sample.Metric.Name,
				Value:    sample.Value,
				Tags:     sample.Tags.Map(),
				Metadata: sample.Metadata,
			}); err != nil {
				o.logger.WithError(err).Error("error appending row to batch")
				return
			}
		}
	}

	if err := batch.Send(); err != nil {
		o.logger.WithError(err).Error("error sending batch")
		return
	}

	o.logger.
		WithFields(map[string]interface{}{
			"duration": time.Since(start),
			"count":    count,
		}).
		Debug("emitted samples")
}
