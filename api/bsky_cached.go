package api

import (
	"context"
	"fmt"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"sync"
	"time"
)

type cachedBlueSkyClient struct {
	renewalThreshold time.Duration
	creds            *bluesky.Credentials

	mu sync.Mutex

	fetchedAt    time.Time
	cachedClient *bluesky.Client
}

func (cb *cachedBlueSkyClient) Get(ctx context.Context) (c *bluesky.Client, err error) {
	ctx, span := tracer.Start(ctx, "cachedBlueSkyClient.Get")
	defer func() {
		endSpan(span, err)
	}()
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// If client was created within the last five minutes, return that client.
	if time.Since(cb.fetchedAt) < cb.renewalThreshold {
		if cb.cachedClient != nil {
			span.AddEvent("client created within last five minutes, returning this client.")
			return cb.cachedClient, nil
		}
	}
	span.AddEvent("no client created within last five minutes, will attempt to create new client.")

	// Otherwise return a new client.
	// TODO: DI PDSHost in or tests wont work
	c, err = bluesky.ClientFromCredentials(ctx, bluesky.DefaultPDSHost, cb.creds)
	if err != nil {
		return nil, fmt.Errorf("fetching token from credentials: %w", err)
	}
	cb.cachedClient = c
	cb.fetchedAt = time.Now()

	return c, nil
}
