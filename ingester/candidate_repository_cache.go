package ingester

import (
	"context"
	"fmt"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"sync"
	"time"
)

type CandidateRepositoryCache struct {
	queries *store.Queries
	cached  map[string]bff.CandidateRepository
	mu      sync.RWMutex
	log     *zap.Logger
}

func NewCandidateRepositoryCache(
	log *zap.Logger,
	queries *store.Queries,
) *CandidateRepositoryCache {
	return &CandidateRepositoryCache{
		queries: queries,
		log:     log,
	}
}

func (crc *CandidateRepositoryCache) GetByDID(
	did string,
) *bff.CandidateRepository {
	crc.mu.RLock()
	defer crc.mu.RUnlock()
	v, ok := crc.cached[did]
	if ok {
		return &v
	}
	return nil
}

func (crc *CandidateRepositoryCache) Fill(ctx context.Context) error {
	crc.log.Info("starting cache fill")
	data, err := crc.queries.ListCandidateRepositories(ctx)
	if err != nil {
		return fmt.Errorf("listing candidate repositories: %w", err)
	}

	mapped := map[string]bff.CandidateRepository{}
	for _, cr := range data {
		mapped[cr.DID] = bff.CandidateRepositoryFromStore(cr)
	}

	crc.mu.Lock()
	defer crc.mu.Unlock()
	crc.cached = mapped
	crc.log.Info("finished cache fill", zap.Int("count", len(mapped)))
	return nil
}

func (crc *CandidateRepositoryCache) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			if err := crc.Fill(ctx); err != nil {
				crc.log.Error("failed to fill cache", zap.Error(err))
			}
			cancel()
		}
	}
}
