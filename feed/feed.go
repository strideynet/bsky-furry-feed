package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/tristate"
)

var feedRequestMetric = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "bff_feed_request_duration_seconds",
	Help: "A very rudimentary way of tracking how many feed skeletons have been requested and how long it takes to serve.",
}, []string{"feed_name", "status"})

type Meta struct {
	// ID is the rkey that is used to identify the Feed in generation requests.
	ID string
	// DisplayName is the short name of the feed used in the BlueSky client.
	DisplayName string
	// Description is a long description of the feed used in the BlueSky client.
	Description string

	// Priority controls where the feed shows up on FurryList UIs.
	// Higher priority wins. Negative values indicate the feed should be hidden
	// in the UI.
	Priority int32
	// TODO: Categories
	// TODO: "Parents"
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

func (s *Service) Metas() []Meta {
	metas := make([]Meta, 0, len(s.feeds))
	for _, f := range s.feeds {
		metas = append(metas, f.meta)
	}
	return metas
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

type generatorOpts struct {
	Hashtags           []string
	DisallowedHashtags []string
	IsNSFW             tristate.Tristate
	HasMedia           tristate.Tristate
}

type chronologicalGeneratorOpts struct {
	generatorOpts
	PinnedDIDs []string
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
			Limit:              limit,
			Hashtags:           opts.Hashtags,
			DisallowedHashtags: opts.DisallowedHashtags,
			IsNSFW:             opts.IsNSFW,
			HasMedia:           opts.HasMedia,
			PinnedDIDs:         opts.PinnedDIDs,
			CursorTime:         cursorTime,
		}

		storePosts, err := pgxStore.ListPostsForNewFeed(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("executing ListPostsForNewFeed: %w", err)
		}

		posts := make([]Post, 0, len(storePosts))
		for _, p := range storePosts {
			posts = append(posts, Post{
				URI:    p.URI,
				Cursor: bluesky.FormatTime(p.IndexedAt.Time),
			})
		}

		return posts, nil
	}
}

type preScoredGeneratorOpts struct {
	generatorOpts
	Alg string
}

func preScoredGenerator(opts preScoredGeneratorOpts) GenerateFunc {
	return func(ctx context.Context, pgxStore *store.PGXStore, cursor string, limit int) ([]Post, error) {
		type cursorValues struct {
			GenerationSeq int64   `json:"generation_seq"`
			AfterScore    float32 `json:"after_score"`
			AfterURI      string  `json:"after_uri"`
		}
		params := store.ListPostsForHotFeedOpts{
			Limit:              limit,
			Hashtags:           opts.Hashtags,
			DisallowedHashtags: opts.DisallowedHashtags,
			IsNSFW:             opts.IsNSFW,
			HasMedia:           opts.HasMedia,
			Alg:                opts.Alg,
		}
		if cursor == "" {
			seq, err := pgxStore.GetLatestScoreGeneration(ctx, opts.Alg)
			if err != nil {
				return nil, fmt.Errorf("executing GetLatestScoreGeneration: %w", err)
			}
			params.Cursor = store.ListPostsForHotFeedCursor{
				GenerationSeq: seq,
				AfterScore:    float32(math.Inf(1)),
				AfterURI:      "",
			}
		} else {
			var p cursorValues
			if err := json.Unmarshal([]byte(cursor), &p); err != nil {
				return nil, fmt.Errorf("unmarshaling cursor: %w", err)
			}
			params.Cursor = store.ListPostsForHotFeedCursor{
				GenerationSeq: p.GenerationSeq,
				AfterScore:    p.AfterScore,
				AfterURI:      p.AfterURI,
			}
		}
		storePosts, err := pgxStore.ListScoredPosts(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("executing ListPostsForHotFeed: %w", err)
		}

		posts := make([]Post, 0, len(storePosts))
		for _, p := range storePosts {
			postCursor, err := json.Marshal(cursorValues{
				GenerationSeq: params.Cursor.GenerationSeq,
				AfterScore:    p.Score,
				AfterURI:      p.URI,
			})
			if err != nil {
				return nil, fmt.Errorf("marshaling cursor: %w", err)
			}
			posts = append(posts, Post{
				URI:    p.URI,
				Cursor: string(postCursor),
			})
		}

		return posts, nil
	}
}

var defaultDisallowedHashtags = []string{"ai", "aiart", "aiartist", "aigenerated", "stablediffusion", "sdxl"}

// ServiceWithDefaultFeeds instantiates a registry with all the standard
// bksy-furry-feed feeds.
// TODO: This really doesn't belong here, ideally, these feeds would be defined
// elsewhere to make this more pluggable. A refactor idea for the future :)
func ServiceWithDefaultFeeds(pgxStore *store.PGXStore) *Service {
	r := &Service{
		store: pgxStore,
	}

	// Hot based feeds
	r.Register(Meta{
		ID:          "furry-hot",
		DisplayName: "🐾 Hot",
		Description: "Hottest posts by furries across Bluesky. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
		Priority:    100,
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "hot-nsfw",
		DisplayName: "🐾 Hot 🌙",
		Description: "Hottest NSFW posts by furries across Bluesky. Contains only NSFW content.\n\nJoin the furry feeds by following @furryli.st",
		Priority:    100,
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			IsNSFW:             tristate.True,
		},
	}))

	// Reverse chronological based feeds
	r.Register(Meta{
		ID:          "furry-new",
		DisplayName: "🐾 New",
		Description: "Posts by furries across Bluesky. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
		Priority:    101,
	}, chronologicalGenerator(chronologicalGeneratorOpts{}))
	r.Register(Meta{
		ID:          "furry-fursuit",
		DisplayName: "🐾 Fursuits",
		Description: "Posts by furries with #fursuit.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursuit", "fursuitfriday"},
			DisallowedHashtags: defaultDisallowedHashtags,
			HasMedia:           tristate.True,
		},
	},
	))
	r.Register(Meta{
		ID:          "fursuit-nsfw",
		DisplayName: "🐾 Murrsuits 🌙",
		Description: "Posts by furries that have an image and #murrsuit or #fursuit.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursuit", "fursuitfriday", "murrsuit", "mursuit"},
			DisallowedHashtags: defaultDisallowedHashtags,
			HasMedia:           tristate.True,
			IsNSFW:             tristate.True,
		},
	},
	))
	r.Register(Meta{
		ID:          "fursuit-clean",
		DisplayName: "🐾 Fursuits 🧼",
		Description: "Posts by furries with #fursuit that haven't been marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursuit", "fursuitfriday"},
			DisallowedHashtags: defaultDisallowedHashtags,
			HasMedia:           tristate.True,
			IsNSFW:             tristate.False,
		},
	},
	))
	r.Register(Meta{
		ID:          "furry-art",
		DisplayName: "🐾 Art",
		Description: "Posts by furries with #furryart. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"furryart"},
			DisallowedHashtags: defaultDisallowedHashtags,
			HasMedia:           tristate.True,
		},
	},
	))
	r.Register(Meta{
		ID:          "art-clean",
		DisplayName: "🐾 Art 🧼",
		Description: "Posts by furries with #furryart and that haven't been marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"furryart"},
			DisallowedHashtags: defaultDisallowedHashtags,
			HasMedia:           tristate.True,
			IsNSFW:             tristate.False,
		},
	}))
	r.Register(Meta{
		ID:          "art-nsfw",
		DisplayName: "🐾 Art 🌙",
		Description: "Posts by furries with #furryart and marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"furryart"},
			DisallowedHashtags: defaultDisallowedHashtags,
			HasMedia:           tristate.True,
			IsNSFW:             tristate.True,
		},
	}))
	r.Register(Meta{
		ID:          "furry-nsfw",
		DisplayName: "🐾 New 🌙",
		Description: "Posts by furries that have been marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			IsNSFW:             tristate.True,
		},
	}))
	r.Register(Meta{
		ID:          "furry-comms",
		DisplayName: "🐾 #CommsOpen",
		Description: "Posts by furries that have #commsopen.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"commsopen"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-denfur",
		DisplayName: "🐾 DenFur 2023",
		Description: "A feed for all things DenFur! Use #denfur or #denfur2023 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"denfur", "denfur2023"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-eurofurence",
		DisplayName: "🐾 Eurofurence 2023",
		Description: "A feed for all things Eurofurence! Use #eurofurence, #eurofurence2023, #eurofurence27, #ef, #ef2023, or #ef27 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"eurofurence", "eurofurence2023", "eurofurence27", "ef", "ef2023", "ef27",

				// I typoed this like 5 times while making this feed so I'm adding these corrections
				"euroference", "euroference2023", "euroference27",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-blfc",
		DisplayName: "🐾 BLFC 2023",
		Description: "A feed for all things BLFC! Use #blfc, #blfc23, or #blfc2023 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"blfc", "blfc23", "blfc2023",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-mff",
		DisplayName: "🐾 MFF 2023",
		Description: "A feed for all things MFF! Use #furfest, #mff, #mff23, or #mff2023 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"furfest", "furfest23", "furfest2023", "mff", "mff23", "mff2023",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-fc",
		DisplayName: "🐾 FC 2024",
		Description: "A feed for all things FC! Use #fc, #fc24, or #fc2024 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"fc", "fc24", "fc2024",
				"furcon", "furcon24", "furcon2024",
				"furtherconfusion", "furtherconfusion24", "furtherconfusion2024",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "merch",
		DisplayName: "🐾 #FurSale",
		Description: "Buy and sell furry merch on the FurSale feed. Use #fursale or #merch to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursale", "merch"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "streamers",
		DisplayName: "🐾 Streamers",
		Description: "Find furs going live on streaming platforms. Use #goinglive or #furrylive to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"goinglive", "furrylive"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "games",
		DisplayName: "🐾 Games",
		Description: "A feed for talking about and showing off furry visual novels and games. Use #FurryVN or #FurryGame to include a post in the feed. \n\nSponsored by @MinoHotel.bsky.social\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"furryvn", "furrygames", "furrygame"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "literature",
		DisplayName: "🐾 Literature",
		Description: "A feed for talking about and showing off furry literature. Use #FurFic, #FurLit or #FurryWriting to include a post in the feed. \n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"furfic",
				"furlit",
				"furrylit",
				"furryfic",
				"furryfiction",
				"furryliterature",
				"furrywriting",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "furry-test",
		DisplayName: "🐾 Test 🚨🛠️",
		Description: "Experimental version of the '🐾 Hot' feed.\ntest\ntest\n\ndouble break",
		Priority:    -1,
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))

	return r
}
