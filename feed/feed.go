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
	"golang.org/x/exp/slices"
	"math"
	"strings"
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

type Post struct {
	URI    string
	Cursor string
}

type GenerateFunc func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]Post, error)

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

	return f.generate(ctx, s.queries, cursor, limit)
}
func PostsFromStorePosts(storePosts []store.CandidatePost) []Post {
	posts := make([]Post, 0, len(storePosts))
	for _, p := range storePosts {
		posts = append(posts, Post{
			URI:    p.URI,
			Cursor: bluesky.FormatTime(p.CreatedAt.Time),
		})
	}
	return posts
}

func newGenerator() GenerateFunc {
	return func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]Post, error) {
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
		return PostsFromStorePosts(posts), nil
	}
}

func newWithTagGenerator(tag string) GenerateFunc {
	return func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]Post, error) {
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
		return PostsFromStorePosts(posts), nil
	}
}

func scoreBasedGenerator(gravity float64, postAgeOffset time.Duration) GenerateFunc {
	return func(ctx context.Context, queries *store.Queries, cursor string, limit int) ([]Post, error) {
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

		rows, err := queries.GetPostsWithLikes(ctx, store.GetPostsWithLikesParams{
			Limit: 2000,
			CursorTimestamp: pgtype.Timestamptz{
				Time:  cursorTime,
				Valid: true,
			},
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
func ServiceWithDefaultFeeds(queries *store.Queries) *Service {
	r := &Service{
		queries: queries,
	}

	r.Register(Meta{ID: "furry-new"}, newGenerator())
	r.Register(Meta{ID: "furry-hot"}, scoreBasedGenerator(1.85, time.Hour*2))
	r.Register(Meta{ID: "furry-fursuit"}, newWithTagGenerator(bff.TagFursuitMedia))
	r.Register(Meta{ID: "furry-art"}, newWithTagGenerator(bff.TagArt))
	r.Register(Meta{ID: "furry-nsfw"}, newWithTagGenerator(bff.TagNSFW))
	r.Register(Meta{ID: "furry-test"}, func(_ context.Context, _ *store.Queries, _ string, limit int) ([]Post, error) {
		return []Post{
			{
				URI: "at://did:plc:dllwm3fafh66ktjofzxhylwk/app.bsky.feed.post/3jznh32lq6s2c",
			},
		}, nil
	})

	return r
}
