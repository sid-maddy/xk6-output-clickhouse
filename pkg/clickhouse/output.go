package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
	"go.uber.org/zap"

	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse/config"
	"github.com/sid-maddy/xk6-output-clickhouse/pkg/clickhouse/logger"
)

// Output implements the [output.Output] interface.
//
//nolint:govet // fieldalignment contradicts with embeddedstructfieldcheck
type Output struct {
	output.SampleBuffer

	conn            clickhouse.Conn
	periodicFlusher *output.PeriodicFlusher

	config *config.Config
	logger *zap.Logger
}

var _ output.Output = new(Output)

// outputRow represents a row of the run metrics.
//
// NOTE: The fields are in snake_case to match the column names in the ClickHouse table.
// NOTE: The field order is different from the ClickHouse table to minimize memory footprint.
type outputRow struct {
	Time time.Time `ch:"time"`

	Tags     map[string]string `ch:"tags"`
	Metadata map[string]string `ch:"metadata"`

	AccountID string  `ch:"account_id"`
	RunID     string  `ch:"run_id"`
	Region    string  `ch:"region"`
	Metric    string  `ch:"metric"`
	Value     float64 `ch:"value"`
}

// New creates an instance of the extension.
func New(params *output.Params) (*Output, error) {
	cfg, err := config.New(params)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	log, err := logger.New(cfg, params.StdOut)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	log.Debug("Opening connection to ClickHouse")

	conn, err := clickhouse.Open(cfg.ClientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to ClickHouse: %w", err)
	}

	//nolint:exhaustruct // periodicFlusher is set later.
	return &Output{
		conn: conn,

		config: cfg,
		logger: log,
	}, nil
}

// Description returns a human-readable description of the output that will be shown in `k6 run`.
func (o *Output) Description() string {
	return fmt.Sprintf("clickhouse (%s)", o.config.DSN)
}

// Start performs initialization tasks prior to the engine using the output.
func (o *Output) Start() error {
	o.logger.Debug("Starting")

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultConnectionVerificationTimeout)
	defer cancel()

	database := o.config.ClientOptions.Auth.Database
	log := o.logger.WithLazy(zap.String("db", database))
	log.Debug("Verifying database")

	if err := o.conn.Exec(ctx, "USE "+database); err != nil {
		o.logger.Error("Database does not exist", zap.Error(err))

		return fmt.Errorf("error verifying database: %w", err)
	}

	o.logger.Debug("Verifying table")

	if err := o.conn.Exec(ctx, fmt.Sprintf("DESCRIBE TABLE %s.%s", database, o.config.Table)); err != nil {
		o.logger.Error("Table does not exist", zap.String("table", o.config.Table), zap.Error(err))

		return fmt.Errorf("error verifying table: %w", err)
	}

	pushInterval := o.config.PushInterval.TimeDuration()
	o.logger.Debug("Creating periodic flusher", zap.Duration("pushInterval", pushInterval))

	periodicFlusher, err := output.NewPeriodicFlusher(pushInterval, o.flushMetrics)
	if err != nil {
		return fmt.Errorf("error creating periodic flusher: %w", err)
	}

	o.periodicFlusher = periodicFlusher

	o.logger.Debug("Started")

	return nil
}

// Stop flushes all remaining metrics and finalizes the test run.
func (o *Output) Stop() error {
	o.logger.Debug("Stopping")

	o.periodicFlusher.Stop()

	if err := o.conn.Close(); err != nil {
		return fmt.Errorf("error closing ClickHouse connection: %w", err)
	}

	if err := o.logger.Sync(); err != nil {
		return fmt.Errorf("error flushing logs: %w", err)
	}

	o.logger.Debug("Stopped")

	return nil
}

// flushMetrics periodically flushes buffered metric samples to ClickHouse.
func (o *Output) flushMetrics() {
	samples := o.GetBufferedSamples()
	if len(samples) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.config.PushInterval.TimeDuration())
	defer cancel()

	batch, err := prepareBatch(ctx, o.logger, o.conn, o.config.ClientOptions.Auth.Database, o.config.Table)
	if err != nil {
		return
	}

	if err := o.emit(ctx, batch, samples); err != nil {
		o.logger.Error("Failed to emit samples to ClickHouse", zap.Error(err))

		return
	}
}

func (o *Output) emit(ctx context.Context, batch *Batch, sampleContainers []metrics.SampleContainer) error {
	start := time.Now()

	o.logger.Debug("Emitting samples to ClickHouse")

	var count int

	select {
	case <-ctx.Done():
		return fmt.Errorf("error emitting samples to ClickHouse: %w", ctx.Err())

	default:
		for _, sc := range sampleContainers {
			samples := sc.GetSamples()

			count += len(samples)

			for _, sample := range samples {
				if err := batch.append(&outputRow{
					Time: sample.Time,

					Tags:     sample.Tags.Map(),
					Metadata: sample.Metadata,

					AccountID: o.config.AccountID,
					Region:    o.config.Region,
					RunID:     o.config.RunID,
					Metric:    sample.Metric.Name,
					Value:     sample.Value,
				}); err != nil {
					return err
				}
			}
		}

		if err := batch.send(); err != nil {
			return err
		}

		o.logger.Debug(
			"Emitted samples to ClickHouse",
			zap.Duration("duration", time.Since(start)),
			zap.Int("count", count),
		)

		return nil
	}
}
