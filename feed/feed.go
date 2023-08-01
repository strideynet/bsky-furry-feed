package feed

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"golang.org/x/exp/slices"
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

type Post struct {
	URI    string
	Cursor string
}

type GenerateFunc func(ctx context.Context, queries *store.PGXStore, cursor string, limit int) ([]Post, error)

type Service struct {
	// TODO: Locking on feeds to avoid data races
	feeds map[string]*feed
	store *store.PGXStore
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
	return ids
}

func (s *Service) GetFeedPosts(ctx context.Context, feedKey string, cursor string, limit int) (posts []Post, err error) {
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

	return f.generate(ctx, s.store, cursor, limit)
}

func PostsFromStorePosts(storePosts []gen.CandidatePost) []Post {
	posts := make([]Post, 0, len(storePosts))
	for _, p := range storePosts {
		posts = append(posts, Post{
			URI:    p.URI,
			Cursor: bluesky.FormatTime(p.IndexedAt.Time),
		})
	}
	return posts
}

type chronologicalGeneratorOpts struct {
	IncludeHashtags []string
	ExcludeHashtags []string
	HasMedia        pgtype.Bool
}

func chronologicalGenerator(opts chronologicalGeneratorOpts) GenerateFunc {
	return func(ctx context.Context, pgxStore *store.PGXStore, cursor string, limit int) ([]Post, error) {
		cursorTime := time.Now().UTC()
		if cursor != "" {
			parsedTime, err := bluesky.ParseTime(cursor)
			if err != nil {
				return nil, fmt.Errorf("parsing cursor: %w", err)
			}
			cursorTime = parsedTime
		}
		params := store.ListPostsForNewFeedOpts{
			Limit:           limit,
			IncludeHashtags: opts.IncludeHashtags,
			ExcludeHashtags: opts.ExcludeHashtags,
			HasMedia:        opts.HasMedia,
			CursorTime:      cursorTime,
		}

		posts, err := pgxStore.ListPostsForNewFeed(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("executing ListPostsForNewFeed: %w", err)
		}
		return PostsFromStorePosts(posts), nil
	}
}

func scoreBasedGenerator(gravity float64, postAgeOffset time.Duration) GenerateFunc {
	return func(ctx context.Context, pgxStore *store.PGXStore, cursor string, limit int) ([]Post, error) {
		cursorTime := time.Now().UTC()
		if cursor != "" {
			parts := strings.Split(cursor, "|")
			if len(parts) != 2 {
				return nil, fmt.Errorf("unexpected number of parts in cursor: %d", len(parts))
			}
			parsedTime, err := bluesky.ParseTime(parts[0])
			if err != nil {
				return nil, fmt.Errorf("parsing cursor time: %w", err)
			}
			cursorTime = parsedTime
		}

		rows, err := pgxStore.ListPostsWithLikes(ctx, store.ListPostsWithLikesOpts{
			Limit:      2000,
			CursorTime: cursorTime,
		})
		if err != nil {
			return nil, fmt.Errorf("executing sql: %w", err)
		}

		type scoredPost struct {
			Post
			Score float64
			Likes int64
			Age   time.Duration
		}

		cursorTimeString := bluesky.FormatTime(cursorTime)
		makeCursor := func(uri string) string {
			return fmt.Sprintf("%s|%s", cursorTimeString, uri)
		}

		scorePost := func(likes int64, age time.Duration) float64 {
			return float64(likes) / math.Pow(age.Hours()+postAgeOffset.Hours(), gravity)
		}

		scoredPosts := make([]scoredPost, 0, len(rows))
		for _, p := range rows {
			age := time.Since(p.IndexedAt.Time)
			scoredPosts = append(scoredPosts, scoredPost{
				Post: Post{
					URI:    p.URI,
					Cursor: makeCursor(p.URI),
				},
				Likes: p.Likes,
				Age:   age,
				Score: scorePost(p.Likes, age),
			})
		}

		slices.SortStableFunc(scoredPosts, func(a, b scoredPost) bool {
			return a.Score > b.Score
		})

		// Strip points info so we can return this in the expected type.
		posts := make([]Post, 0, len(rows))
		for _, p := range scoredPosts {
			// Debugs the post scoring for top ten - we need to add an endpoint
			// for this.
			posts = append(posts, p.Post)
		}

		if cursor != "" {
			// This pagination is extremely rough - we search through the list
			// for the current post in the cursor and then start from there.
			foundIndex := -1
			for i, p := range posts {
				if p.Cursor == cursor {
					foundIndex = i
					break
				}
			}
			if foundIndex == -1 {
				// cant find post, indicate to client to start again
				return nil, fmt.Errorf("could not find cursor post")
			}
			if foundIndex+1 == len(posts) {
				// the cursor is pointing at the last post, indicate end
				// reached by returning empty
				return nil, nil
			}
			posts = posts[foundIndex+1:]
		}

		posts = posts[:limit]

		return posts, nil
	}
}

// ServiceWithDefaultFeeds instantiates a registry with all the standard
// bksy-furry-feed feeds.
// TODO: This really doesn't belong here, ideally, these feeds would be defined
// elsewhere to make this more pluggable. A refactor idea for the future :)
func ServiceWithDefaultFeeds(pgxStore *store.PGXStore) *Service {
	r := &Service{
		store: pgxStore,
	}

	// Hot based feeds
	r.Register(Meta{ID: "furry-hot"}, scoreBasedGenerator(1.85, time.Hour*2))

	// Reverse chronological based feeds
	r.Register(Meta{ID: "furry-new"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{},
		ExcludeHashtags: []string{},
	}))
	r.Register(Meta{ID: "furry-fursuit"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"fursuitfriday", "fursuit", "murrsuit", "mursuit"},
		ExcludeHashtags: []string{},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	}))
	r.Register(Meta{ID: "fursuit-nsfw"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"fursuitfriday", "fursuit", "murrsuit", "mursuit", "nsfw"},
		ExcludeHashtags: []string{},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	}))
	r.Register(Meta{ID: "fursuit-clean"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"fursuitfriday", "fursuit"},
		ExcludeHashtags: []string{"murrsuit", "mursuit", "nsfw"},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	}))
	r.Register(Meta{ID: "furry-art"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"art", "furryart"},
		ExcludeHashtags: []string{},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	}))
	r.Register(Meta{ID: "art-clean"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"art", "furryart"},
		ExcludeHashtags: []string{"nsfw"},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	}))
	r.Register(Meta{ID: "art-nsfw"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"art", "furryart", "nsfw"},
		ExcludeHashtags: []string{},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	}))
	r.Register(Meta{ID: "furry-nsfw"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"nsfw"},
		ExcludeHashtags: []string{},
	}))
	r.Register(Meta{ID: "furry-comms"}, chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"commsopen"},
		ExcludeHashtags: []string{},
	}))
	r.Register(Meta{ID: "furry-test"}, func(_ context.Context, _ *store.PGXStore, _ string, limit int) ([]Post, error) {
		return []Post{
			{
				URI: "at://did:plc:dllwm3fafh66ktjofzxhylwk/app.bsky.feed.post/3jznh32lq6s2c",
			},
		}, nil
	})

	return r
}
