package scoring

import (
	"context"
	"fmt"
	"time"

	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

type Materializer struct {
	log   *zap.Logger
	store *store.PGXStore
	opts  Opts
}

type Opts struct {
	MaterializationInterval time.Duration
	RetentionPeriod         time.Duration
	LookbackPeriod          time.Duration
}

func NewMaterializer(
	log *zap.Logger, store *store.PGXStore, opts Opts,
) *Materializer {
	return &Materializer{
		log:   log,
		store: store,
		opts:  opts,
	}
}

func (m *Materializer) materialize(ctx context.Context) error {
	now := time.Now()
	seq, err := m.store.MaterializeClassicPostScores(ctx, now.Add(-m.opts.LookbackPeriod))
	if err != nil {
		return err
	}
	m.log.Info(
		"materialized generation",
		zap.Int64("seq", seq),
		zap.Duration("duration", time.Since(now)),
	)
	return nil
}

func (m *Materializer) cleanup(ctx context.Context) error {
	now := time.Now()
	n, err := m.store.DeleteOldPostScores(ctx, now.Add(-m.opts.RetentionPeriod))
	if err != nil {
		return err
	}
	m.log.Info("cleaned up old rows", zap.Int64("n", n))
	return nil
}

func (m *Materializer) step(ctx context.Context) error {
	// NOTE: materalize and cleanup don't run a transaction together (they don't need to, since it's okay if we keep old materialized results around for too long).
	// However, we should do the cleanup _after_ the materialization, in case the materialization fails and we converge on purging the entire table.
	if err := m.materialize(ctx); err != nil {
		return fmt.Errorf("materializing: %w", err)
	}
	if err := m.cleanup(ctx); err != nil {
		return fmt.Errorf("cleaning up: %w", err)
	}
	return nil
}

func (m *Materializer) Run(ctx context.Context) error {
	t := time.NewTicker(m.opts.MaterializationInterval)
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return ctx.Err()
		}

		if err := m.step(ctx); err != nil {
			m.log.Error("failed to execute step", zap.Error(err))
		}
	}
}
