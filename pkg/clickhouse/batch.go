package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

var (
	errPreparingBatchInsertQuery = errors.New("error preparing batch insert query")
	errAppendingRowToBatch       = errors.New("error appending row to batch")
	errSendingBatch              = errors.New("error sending batch")
)

// Batch represents a batch insert query.
type Batch struct {
	batch  driver.Batch
	logger *zap.Logger
}

func prepareBatch(
	ctx context.Context,
	logger *zap.Logger,
	conn clickhouse.Conn,
	database, table string,
) (*Batch, error) {
	logger.Debug("Preparing batch insert query", zap.String("database", database), zap.String("table", table))

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %w", errPreparingBatchInsertQuery, ctx.Err())

	default:
		batch, err := conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s", database, table))
		if err != nil {
			logger.Error(strings.ToTitle(errPreparingBatchInsertQuery.Error()), zap.Error(err))

			return nil, fmt.Errorf("%w: %w", errPreparingBatchInsertQuery, err)
		}

		return &Batch{batch: batch, logger: logger}, nil
	}
}

func (b *Batch) append(row any) error {
	b.logger.Debug("Appending row to batch")

	if err := b.batch.AppendStruct(row); err != nil {
		b.logger.Error(strings.ToTitle(errAppendingRowToBatch.Error()), zap.Error(err))

		return fmt.Errorf("%w: %w", errAppendingRowToBatch, err)
	}

	return nil
}

func (b *Batch) send() error {
	b.logger.Debug("Sending batch", zap.Int("count", b.batch.Rows()))

	if err := b.batch.Send(); err != nil {
		b.logger.Error(strings.ToTitle(errSendingBatch.Error()), zap.Error(err))

		return fmt.Errorf("%w: %w", errSendingBatch, err)
	}

	return nil
}
