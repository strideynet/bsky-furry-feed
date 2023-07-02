package feed

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"time"
)

var feedRequestMetric = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "bff_feed_request_duration_seconds",
	Help: "A very rudimentary way of tracking how many feed skeletons have been requested and how long it takes to serve.",
}, []string{"feed_name", "status"})

type Meta struct {
	// ID is the rkey that is used to identify the Feed in generation requests.
	ID string
	// TODO: Add desc/name fields which can then be used in feed upload command.
}

type feed struct {
	meta     Meta
	generate GenerateFunc
}

type GenerateFunc func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]store.CandidatePost, error)

type Service struct {
	// TODO: Locking on feeds to avoid data races
	feeds   map[string]*feed
	queries *store.Queries
}

func (s *Service) Register(m Meta, generateFunc GenerateFunc) {
	if s.feeds == nil {
		s.feeds = map[string]*feed{}
	}
	s.feeds[m.ID] = &feed{
		meta:     m,
		generate: generateFunc,
	}
}

// IDs returns a slice of the IDs of feeds which are eligible for generation.
func (s *Service) IDs() []string {
	ids := make([]string, len(s.feeds))
	for _, f := range s.feeds {
		ids = append(ids, f.meta.ID)
	}
	return nil
}

func (s *Service) GetFeedPosts(ctx context.Context, feedKey string, cursor string, limit int) (posts []store.CandidatePost, err error) {
	start := time.Now()
	defer func() {
		status := "OK"
		if err != nil {
			status = "ERR"
		}
		feedRequestMetric.
			WithLabelValues(feedKey, status).
			Observe(time.Since(start).Seconds())
	}()

	f, ok := s.feeds[feedKey]
	if !ok {
		return nil, fmt.Errorf("unrecognized feed")
	}

	return f.generate(ctx, s.queries, cursor, limit)
}

func newGenerator() GenerateFunc {
	return func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]store.CandidatePost, error) {
		params := store.GetFurryNewFeedParams{
			Limit: int32(limit),
		}
		if cursor != "" {
			cursorTime, err := bluesky.ParseTime(cursor)
			if err != nil {
				return nil, fmt.Errorf("parsing cursor: %w", err)
			}
			params.CursorTimestamp = pgtype.Timestamptz{
				Valid: true,
				Time:  cursorTime,
			}
		}

		posts, err := queries.GetFurryNewFeed(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("executing sql: %w", err)
		}
		return posts, nil
	}
}

func newWithTagGenerator(tag string) GenerateFunc {
	return func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]store.CandidatePost, error) {
		params := store.GetFurryNewFeedWithTagParams{
			Limit: int32(limit),
			Tag:   tag,
		}
		if cursor != "" {
			cursorTime, err := bluesky.ParseTime(cursor)
			if err != nil {
				return nil, fmt.Errorf("parsing cursor: %w", err)
			}
			params.CursorTimestamp = pgtype.Timestamptz{
				Valid: true,
				Time:  cursorTime,
			}
		}

		posts, err := queries.GetFurryNewFeedWithTag(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("executing sql: %w", err)
		}
		return posts, nil
	}
}

func hotGenerator() GenerateFunc {
	return func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]store.CandidatePost, error) {
		params := store.GetFurryHotFeedParams{
			Limit:         int32(limit),
			LikeThreshold: int32(4),
		}
		if cursor != "" {
			cursorTime, err := bluesky.ParseTime(cursor)
			if err != nil {
				return nil, fmt.Errorf("parsing cursor: %w", err)
			}
			params.CursorTimestamp = pgtype.Timestamptz{
				Valid: true,
				Time:  cursorTime,
			}
		}

		posts, err := queries.GetFurryHotFeed(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("executing sql: %w", err)
		}
		return posts, nil
	}
}

// ServiceWithDefaultFeeds instantiates a registry with all the standard
// bksy-furry-feed feeds.
// TODO: This really doesn't belong here, ideally, these feeds would be defined
// elsewhere to make this more pluggable. A refactor idea for the future :)
func ServiceWithDefaultFeeds(queries *store.Queries) *Service {
	r := &Service{
		queries: queries,
	}

	newGen := newGenerator()
	r.Register(Meta{ID: "furry-new"}, newGen)
	r.Register(Meta{ID: "furry-hot"}, hotGenerator())
	r.Register(Meta{ID: "furry-fursuit"}, newWithTagGenerator(bff.TagFursuitMedia))
	r.Register(Meta{ID: "furry-art"}, newWithTagGenerator(bff.TagArt))
	r.Register(Meta{ID: "furry-nsfw"}, newWithTagGenerator(bff.TagNSFW))
	r.Register(Meta{ID: "furry-test"}, newGen)

	return r
}
