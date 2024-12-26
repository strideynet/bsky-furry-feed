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
	AllowedEmbeds      []EmbedType
}

type chronologicalGeneratorOpts struct {
	generatorOpts
	PinnedDIDs []string
}

type EmbedType string

const (
	EmbedNone  EmbedType = "none"
	EmbedImage EmbedType = "image"
	EmbedVideo EmbedType = "video"
)

var allowImageAndVideo = []EmbedType{EmbedImage, EmbedVideo}
var allowVideoOnly = []EmbedType{EmbedVideo}

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
		allowedEmbeds := []string{}
		for _, embed := range opts.AllowedEmbeds {
			allowedEmbeds = append(allowedEmbeds, string(embed))
		}
		params := store.ListPostsForNewFeedOpts{
			Limit:              limit,
			Hashtags:           opts.Hashtags,
			DisallowedHashtags: opts.DisallowedHashtags,
			IsNSFW:             opts.IsNSFW,
			AllowedEmbeds:      allowedEmbeds,
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
		allowedEmbeds := []string{}
		for _, embed := range opts.AllowedEmbeds {
			allowedEmbeds = append(allowedEmbeds, string(embed))
		}
		params := store.ListPostsForHotFeedOpts{
			Limit:              limit,
			Hashtags:           opts.Hashtags,
			DisallowedHashtags: opts.DisallowedHashtags,
			IsNSFW:             opts.IsNSFW,
			AllowedEmbeds:      allowedEmbeds,
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
		Description: "Furry\nHottest posts by furries across Bluesky. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
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
		Description: "Furry\nHottest NSFW posts by furries across Bluesky. Contains only NSFW content.\n\nJoin the furry feeds by following @furryli.st",
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
		Description: "Furry\nPosts by furries across Bluesky. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
		Priority:    101,
	}, chronologicalGenerator(chronologicalGeneratorOpts{}))
	r.Register(Meta{
		ID:          "furry-fursuit",
		DisplayName: "🐾 Fursuits",
		Description: "Furry\nPosts by furries with #fursuit.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursuit", "fursuitfriday"},
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
		},
	},
	))
	r.Register(Meta{
		ID:          "fursuit-nsfw",
		DisplayName: "🐾 Murrsuits 🌙",
		Description: "Furry\nPosts by furries that have an image and #murrsuit or #fursuit.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursuit", "fursuitfriday", "murrsuit", "mursuit"},
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
			IsNSFW:             tristate.True,
		},
	},
	))
	r.Register(Meta{
		ID:          "fursuit-clean",
		DisplayName: "🐾 Fursuits 🧼",
		Description: "Furry\nPosts by furries with #fursuit that haven't been marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursuit", "fursuitfriday"},
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
			IsNSFW:             tristate.False,
		},
	},
	))
	var furryArtHashtags = []string{"furryart"}
	r.Register(Meta{
		ID:          "furry-art",
		DisplayName: "🐾 Art",
		Description: "Furry\nPosts by furries with #furryart. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           furryArtHashtags,
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
		},
	},
	))
	r.Register(Meta{
		ID:          "art-clean",
		DisplayName: "🐾 Art 🧼",
		Description: "Furry\nPosts by furries with #furryart and that haven't been marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           furryArtHashtags,
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
			IsNSFW:             tristate.False,
		},
	}))
	r.Register(Meta{
		ID:          "art-nsfw",
		DisplayName: "🐾 Art 🌙",
		Description: "Furry\nPosts by furries with #furryart and marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           furryArtHashtags,
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
			IsNSFW:             tristate.True,
		},
	}))
	r.Register(Meta{
		ID:          "art-hot",
		DisplayName: "🐾 Hot Art",
		Description: "Furry\nHottest posts by furries with #furryart. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			Hashtags:           furryArtHashtags,
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
		},
	}))
	r.Register(Meta{
		ID:          "art-hot-nsfw",
		DisplayName: "🐾 Hot Art 🌙",
		Description: "Furry\nHottest posts by furries with #furryart and marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			Hashtags:           furryArtHashtags,
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
			IsNSFW:             tristate.True,
		},
	}))
	r.Register(Meta{
		ID:          "furry-nsfw",
		DisplayName: "🐾 New 🌙",
		Description: "Furry\nPosts by furries that have been marked NSFW.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowImageAndVideo,
			IsNSFW:             tristate.True,
		},
	}))
	r.Register(Meta{
		ID:          "furry-comms",
		DisplayName: "🐾 #CommsOpen",
		Description: "Furry\nPosts by furries that have #commsopen.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"commsopen"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-denfur",
		DisplayName: "🐾 DenFur 2024",
		Description: "Furry\nA feed for all things DenFur! Use #denfur or #denfur2024 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"denfur", "denfur2024"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-eurofurence",
		DisplayName: "🐾 Eurofurence 2024",
		Description: "Furry\nA feed for all things Eurofurence! Use #eurofurence, #eurofurence2024, #eurofurence28, #ef, #ef2024, or #ef28 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"eurofurence", "ef",
				"eurofurence2023", "eurofurence27", "ef2023", "ef27",
				"eurofurence2024", "eurofurence28", "ef2024", "ef28",
				// I typoed this like 5 times while making this feed so I'm adding these corrections
				"euroference",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-blfc",
		DisplayName: "🐾 BLFC 2024",
		Description: "Furry\nA feed for all things BLFC! Use #blfc, #blfc24, or #blfc2024 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"blfc", "blfc24", "blfc2024",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-mff",
		DisplayName: "🐾 MFF 2024",
		Description: "Furry\nA feed for all things MFF! Use #furfest, #mff, #mff24, or #mff2024 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"furfest", "furfest24", "furfest2024", "mff", "mff24", "mff2024",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-fc",
		DisplayName: "🐾 FC 2025",
		Description: "Furry\nA feed for all things FC! Use #fc, #fc25, or #fc2025 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"fc", "fc25", "fc2025", "furcon25", "furcon2025",
				"furtherconfusion", "furtherconfusion25", "furtherconfusion2025",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-nfc",
		DisplayName: "🐾 NFC 2025",
		Description: "Furry\nA feed for all things NFC! Use #nfc, #nfc25, or #nfc2025 to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"nfc", "nfc25", "nfc2025",
				"nordicfuzzcon", "nordicfuzzcon25", "nordicfuzzcon2025",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-fwa",
		DisplayName: "🐾 FWA 2024",
		Description: "Furry\nA feed for all things FWA! Use #fwa, #fwa24, #fwa2024, or #furryweekend to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"fwa", "fwa24", "fwa2024", "furryweekend",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "con-ac",
		DisplayName: "🐾 Anthrocon 2024",
		Description: "Furry\nA feed for all things Anthrocon! Use #anthrocon, #anthrocon2024, or #ac to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags: []string{
				"anthrocon", "anthrocon2024", "anthrocon24",
				"ac", "ac2024", "ac24",
			},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "merch",
		DisplayName: "🐾 #FurSale",
		Description: "Furry\nBuy and sell furry merch on the FurSale feed. Use #fursale or #merch to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"fursale", "merch"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "streamers",
		DisplayName: "🐾 Streamers",
		Description: "Furry\nFind furs going live on streaming platforms. Use #goinglive or #furrylive to include a post in the feed.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"goinglive", "furrylive"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "games",
		DisplayName: "🐾 Games",
		Description: "Furry\nA feed for talking about and showing off furry visual novels and games. Use #FurryVN or #FurryGame to include a post in the feed. \n\nSponsored by @MinoHotel.bsky.social\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			Hashtags:           []string{"furryvn", "furrygames", "furrygame"},
			DisallowedHashtags: defaultDisallowedHashtags,
		},
	}))
	r.Register(Meta{
		ID:          "literature",
		DisplayName: "🐾 Literature",
		Description: "Furry\nA feed for talking about and showing off furry literature. Use #FurFic, #FurLit or #FurryWriting to include a post in the feed. \n\nJoin the furry feeds by following @furryli.st",
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

	r.Register(Meta{
		ID:          "video-hot",
		DisplayName: "🐾 Hot videos",
		Description: "Furry\nHottest video posts by furries across Bluesky. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
		Priority:    100,
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowVideoOnly,
		},
	}))
	r.Register(Meta{
		ID:          "video-new",
		DisplayName: "🐾 New videos",
		Description: "Furry\nLatest video posts by furries across Bluesky. Contains a mix of SFW and NSFW content.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowVideoOnly,
		},
	}))
	r.Register(Meta{
		ID:          "video-hot-nsfw",
		DisplayName: "🐾 Hot videos 🌙",
		Description: "Furry\nHottest NSFW video posts by furries across Bluesky. Contains only NSFW content.\n\nJoin the furry feeds by following @furryli.st",
		Priority:    100,
	}, preScoredGenerator(preScoredGeneratorOpts{
		Alg: "classic",
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowVideoOnly,
			IsNSFW:             tristate.True,
		},
	}))
	r.Register(Meta{
		ID:          "video-new-nsfw",
		DisplayName: "🐾 New videos 🌙",
		Description: "Furry\nLatest NSFW video posts by furries across Bluesky. Contains only NSFW content.\n\nJoin the furry feeds by following @furryli.st",
	}, chronologicalGenerator(chronologicalGeneratorOpts{
		generatorOpts: generatorOpts{
			DisallowedHashtags: defaultDisallowedHashtags,
			AllowedEmbeds:      allowVideoOnly,
			IsNSFW:             tristate.True,
		},
	}))

	return r
}
