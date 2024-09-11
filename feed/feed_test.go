package feed

import (
	"context"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/stretchr/testify/require"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/testenv"
	"github.com/strideynet/bsky-furry-feed/tristate"
)

func TestGenerator(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)
	furry := harness.PDS.MustNewUser(t, "furry.tpds")
	pinnedFurry := harness.PDS.MustNewUser(t, "pinnedFurry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		DID:    furry.DID(),
	})
	require.NoError(t, err)
	_, err = harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		DID:    pinnedFurry.DID(),
	})
	require.NoError(t, err)

	fursuitPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "fursuit")
	murrsuitPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "murrsuit")
	artPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "art")
	nsfwArtPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "nsfwArt")
	poastPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "poast")
	nsfwLabelledPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "nsfw-labelled")
	aiArtPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "ai")
	pinnedPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "pinned-post")
	videoPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "video")
	nsfwVideoPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "nsfw-video")
	nsfwArtVideoPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "nsfw-art-video")
	artVideoPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "art-video")
	oldPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "old-post")

	now := time.Now()

	for _, opts := range []store.CreatePostOpts{
		{
			URI:      fursuitPost,
			Hashtags: []string{"fursuit"},
			HasMedia: true,
		},
		{
			URI:       oldPost,
			ActorDID:  furry.DID(),
			IndexedAt: now.Add(-time.Hour * 24 * 8),
			HasMedia:  true,
		},
		{
			URI:      murrsuitPost,
			Hashtags: []string{"fursuit", "murrsuit"},
			HasMedia: true,
		},
		{
			URI:      artPost,
			Hashtags: []string{"art"},
			HasMedia: true,
		},
		{
			URI:      nsfwArtPost,
			Hashtags: []string{"furryart", "nsfw"},
			HasMedia: true,
		},
		{
			URI:      poastPost,
			HasMedia: true,
		},
		{
			URI:        nsfwLabelledPost,
			Hashtags:   []string{"art"},
			HasMedia:   true,
			SelfLabels: []string{"sexual"},
		},
		{
			URI:        nsfwVideoPost,
			HasVideo:   true,
			Hashtags:   []string{},
			SelfLabels: []string{"sexual"},
		},
		{
			URI:        nsfwArtVideoPost,
			HasVideo:   true,
			Hashtags:   []string{"furryart"},
			SelfLabels: []string{"sexual"},
		},
		{
			URI:      videoPost,
			HasVideo: true,
		},
		{
			URI:      artVideoPost,
			HasVideo: true,
			Hashtags: []string{"art"},
		},
		{
			URI:      aiArtPost,
			Hashtags: []string{"art", "aiart", "aiartist"},
			HasMedia: true,
		},
		{
			URI:      pinnedPost,
			ActorDID: pinnedFurry.DID(),
			HasMedia: true,
		},
	} {
		if opts.ActorDID == "" {
			opts.ActorDID = furry.DID()
		}
		if opts.Raw == nil {
			opts.Raw = &bsky.FeedPost{}
		}
		if opts.Hashtags == nil {
			opts.Hashtags = []string{}
		}
		if opts.IndexedAt.IsZero() {
			opts.IndexedAt = now
		}
		require.NoError(t, harness.Store.CreatePost(ctx, opts))
	}

	t.Run("chronological", func(t *testing.T) {
		t.Parallel()

		for _, test := range []struct {
			name          string
			opts          chronologicalGeneratorOpts
			expectedPosts []string
		}{
			{
				name: "all",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags: []string{},
						IsNSFW:   tristate.Maybe,
						HasMedia: tristate.Maybe,
						HasVideo: tristate.Maybe,
					},
				},
				expectedPosts: []string{
					fursuitPost,
					murrsuitPost,
					artPost,
					nsfwArtPost,
					poastPost,
					nsfwLabelledPost,
					pinnedPost,
					aiArtPost,
					videoPost,
					nsfwVideoPost,
					artVideoPost,
					nsfwArtVideoPost,
				},
			},
			{
				name: "all fursuits",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags: []string{"fursuit"},
						IsNSFW:   tristate.Maybe,
						HasMedia: tristate.True,
					},
				},
				expectedPosts: []string{fursuitPost, murrsuitPost},
			},
			{
				name: "sfw only fursuits",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags: []string{"fursuit"},
						IsNSFW:   tristate.False,
						HasMedia: tristate.True,
					},
				},
				expectedPosts: []string{fursuitPost},
			},
			{
				name: "art",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags:           []string{"art", "furryart"},
						DisallowedHashtags: []string{"aiart"},
						HasMedia:           tristate.True,
					},
				},
				expectedPosts: []string{artPost, nsfwArtPost, nsfwLabelledPost, artVideoPost, nsfwArtVideoPost},
			},
			{
				name: "nsfw only art",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags: []string{"art", "furryart"},
						IsNSFW:   tristate.True,
						HasMedia: tristate.True,
					},
				},
				expectedPosts: []string{nsfwArtPost, nsfwLabelledPost, nsfwArtVideoPost},
			},
			{
				name: "pinned post",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags: []string{"placeholder"},
						HasMedia: tristate.False,
					},
					PinnedDIDs: []string{pinnedFurry.DID()},
				},
				expectedPosts: []string{pinnedPost},
			},
			{
				name: "all videos",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						Hashtags: []string{},
						HasVideo: tristate.True,
						HasMedia: tristate.False,
					},
				},
				expectedPosts: []string{videoPost, nsfwVideoPost, artVideoPost, nsfwArtVideoPost},
			},
			{
				name: "videos nsfw",
				opts: chronologicalGeneratorOpts{
					generatorOpts: generatorOpts{
						IsNSFW:   tristate.True,
						HasVideo: tristate.True,
						HasMedia: tristate.False,
					},
				},
				expectedPosts: []string{nsfwVideoPost, nsfwArtVideoPost},
			},
		} {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				posts, err := chronologicalGenerator(test.opts)(ctx, harness.Store, "", 1000)
				require.NoError(t, err)
				postURIs := make([]string, len(posts))
				for i, post := range posts {
					postURIs[i] = post.URI
				}
				require.ElementsMatch(t, test.expectedPosts, postURIs)
			})
		}
	})

	t.Run("prescored", func(t *testing.T) {
		t.Parallel()

		_, err = harness.Store.MaterializeClassicPostScores(ctx, time.Time{})
		require.NoError(t, err)

		for _, test := range []struct {
			name          string
			opts          preScoredGeneratorOpts
			expectedPosts []string
		}{
			{
				name: "all",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags: []string{},
						IsNSFW:   tristate.Maybe,
						HasMedia: tristate.Maybe,
						HasVideo: tristate.Maybe,
					},
				},
				expectedPosts: []string{
					fursuitPost,
					murrsuitPost,
					artPost,
					nsfwArtPost,
					poastPost,
					nsfwLabelledPost,
					aiArtPost,
					pinnedPost,
					videoPost,
					nsfwVideoPost,
					nsfwArtVideoPost,
					artVideoPost,
				},
			},
			{
				name: "all fursuits",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags: []string{"fursuit"},
						IsNSFW:   tristate.Maybe,
						HasMedia: tristate.True,
					},
				},
				expectedPosts: []string{fursuitPost, murrsuitPost},
			},
			{
				name: "sfw only fursuits",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags: []string{"fursuit"},
						IsNSFW:   tristate.False,
						HasMedia: tristate.True,
					},
				},
				expectedPosts: []string{fursuitPost},
			},
			{
				name: "art",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags:           []string{"art", "furryart"},
						DisallowedHashtags: []string{"aiart"},
						HasMedia:           tristate.True,
					},
				},
				expectedPosts: []string{artPost, nsfwArtPost, nsfwLabelledPost, artVideoPost, nsfwArtVideoPost},
			},
			{
				name: "nsfw only art",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags: []string{"art", "furryart"},
						IsNSFW:   tristate.True,
						HasMedia: tristate.True,
					},
				},
				expectedPosts: []string{nsfwArtPost, nsfwLabelledPost, nsfwArtVideoPost},
			},
			{
				name: "all videos",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags: []string{},
						HasVideo: tristate.True,
						HasMedia: tristate.False,
					},
				},
				expectedPosts: []string{videoPost, nsfwVideoPost, artVideoPost, nsfwArtVideoPost},
			},
			{
				name: "nsfw videos",
				opts: preScoredGeneratorOpts{
					Alg: "classic",
					generatorOpts: generatorOpts{
						Hashtags: []string{},
						IsNSFW:   tristate.True,
						HasVideo: tristate.True,
						HasMedia: tristate.False,
					},
				},
				expectedPosts: []string{nsfwVideoPost, nsfwArtVideoPost},
			},
		} {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				posts, err := preScoredGenerator(test.opts)(ctx, harness.Store, "", 1000)
				require.NoError(t, err)
				postURIs := make([]string, len(posts))
				for i, post := range posts {
					postURIs[i] = post.URI
				}
				require.ElementsMatch(t, test.expectedPosts, postURIs)
			})
		}
	})
}
