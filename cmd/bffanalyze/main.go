package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
)

func main() {
	if err := run(); err != nil {
		slog.Error("exited with error", "error", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	st, err := store.ConnectPGXStore(ctx, zap.L(), &store.DirectConnector{
		URI: os.Getenv("DB_URI"),
	})
	if err != nil {
		return fmt.Errorf("connecting to store: %w", err)
	}
	defer st.Close()

	pool := st.RawPool()
	c, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer c.Release()

	// Run a prepare to see what types are "inferred"
	prepared, err := c.Conn().Prepare(ctx, "type inference", gen.GetFurryNewFeed)
	if err != nil {
		return fmt.Errorf("preparing query: %w", err)
	}
	slog.Info("Ran statement prepare using PGX, the parameters were inferred")
	for _, oid := range prepared.ParamOIDs {
		t, ok := c.Conn().TypeMap().TypeForOID(oid)
		if !ok {
			slog.Info("Unknown oid", "oid", oid)
			continue
		}
		slog.Info("Parameter", "oid", oid, "type", t.Name)
	}

	now := time.Now()
	compare := map[string]gen.GetFurryNewFeedParams{
		"actual_new_feed": gen.GetFurryNewFeedParams{
			Hashtags:           nil,
			DisallowedHashtags: nil,
			IsNSFW:             pgtype.Bool{},
			PinnedDIDs:         nil,
			CursorTimestamp: pgtype.Timestamptz{
				Valid: true,
				Time:  now,
			},
			Limit:         30,
			AllowedEmbeds: []string{},
		},
		"empty_slice": gen.GetFurryNewFeedParams{
			Hashtags:           []string{},
			DisallowedHashtags: []string{},
			IsNSFW:             pgtype.Bool{},
			PinnedDIDs:         []string{},
			CursorTimestamp: pgtype.Timestamptz{
				Valid: true,
				Time:  now,
			},
			Limit:         30,
			AllowedEmbeds: []string{},
		},
	}

	for name, data := range compare {
		// https://explain.dalibo.com/
		res, err := c.Query(
			ctx,
			"EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON) "+gen.GetFurryNewFeed,
			data.Hashtags,
			data.DisallowedHashtags,
			data.IsNSFW,
			data.PinnedDIDs,
			data.CursorTimestamp,
			data.Limit,
			data.AllowedEmbeds,
		)
		if err != nil {
			return fmt.Errorf("querying: %w", err)
		}
		out := []byte{}
		for res.Next() {
			if res.Scan(&out); err != nil {
				return fmt.Errorf("scanning row: %w", err)
			}
		}
		if err := res.Err(); err != nil {
			return fmt.Errorf("iterating rows: %w", err)
		}
		err = os.WriteFile(fmt.Sprintf("analyze_%s.json", name), out, 0644)
		if err != nil {
			return fmt.Errorf("writing plan: %w", err)
		}
		err = os.WriteFile(fmt.Sprintf("analyze_%s.sql", name), []byte(gen.GetFurryNewFeed), 0644)
		if err != nil {
			return fmt.Errorf("writing plan query: %w", err)
		}

		// Speed test
		start := time.Now()
		for i := 0; i < 100; i++ {
			res, err := c.Query(
				ctx, gen.GetFurryNewFeed,
				data.Hashtags,
				data.DisallowedHashtags,
				data.IsNSFW,
				data.PinnedDIDs,
				data.CursorTimestamp,
				data.Limit,
				data.AllowedEmbeds,
			)
			if err != nil {
				return fmt.Errorf("querying: %w", err)
			}
			res.Close()
		}
		stop := time.Now()
		slog.Info("Ran the query 100x", "name", name, "duration", stop.Sub(start))
	}

	return nil
}
