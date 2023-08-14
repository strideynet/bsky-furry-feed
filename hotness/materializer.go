package hotness

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
	MaterializationPeriod time.Duration
	RetentionPeriod       time.Duration
	LookbackPeriod        time.Duration
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
	seq, err := m.store.MaterializeClassicPostHotness(ctx, m.opts.LookbackPeriod)
	if err != nil {
		return err
	}
	m.log.Info("materialized generation", zap.Int64("seq", seq))
	return nil
}

func (m *Materializer) cleanup(ctx context.Context) error {
	n, err := m.store.DeleteOldPostHotness(ctx, m.opts.RetentionPeriod)
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
		return fmt.Errorf("materialize: %w", err)
	}
	if err := m.cleanup(ctx); err != nil {
		return fmt.Errorf("cleanup: %w", err)
	}
	return nil
}

func (m *Materializer) Run(ctx context.Context) error {
	t := time.NewTicker(m.opts.MaterializationPeriod)
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return ctx.Err()
		}

		if err := m.step(ctx); err != nil {
			m.log.Error("step", zap.Error(err))
		}
	}
}
