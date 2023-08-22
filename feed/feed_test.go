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

func TestChronologicalGenerator(t *testing.T) {
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
	pinnedPost := indigoTest.RandFakeAtUri("app.bsky.feed.post", "pinned-post")

	for _, opts := range []store.CreatePostOpts{
		{
			URI:       fursuitPost,
			ActorDID:  furry.DID(),
			CreatedAt: time.Time{},
			IndexedAt: time.Time{},
			Hashtags:  []string{"fursuit"},
			HasMedia:  true,
			Raw:       &bsky.FeedPost{},
		},
		{
			URI:       murrsuitPost,
			ActorDID:  furry.DID(),
			CreatedAt: time.Time{},
			IndexedAt: time.Time{},
			Hashtags:  []string{"fursuit", "murrsuit"},
			HasMedia:  true,
			Raw:       &bsky.FeedPost{},
		},
		{
			URI:       artPost,
			ActorDID:  furry.DID(),
			CreatedAt: time.Time{},
			IndexedAt: time.Time{},
			Hashtags:  []string{"art"},
			HasMedia:  true,
			Raw:       &bsky.FeedPost{},
		},
		{
			URI:       nsfwArtPost,
			ActorDID:  furry.DID(),
			CreatedAt: time.Time{},
			IndexedAt: time.Time{},
			Hashtags:  []string{"furryart", "nsfw"},
			HasMedia:  true,
			Raw:       &bsky.FeedPost{},
		},
		{
			URI:       poastPost,
			ActorDID:  furry.DID(),
			CreatedAt: time.Time{},
			IndexedAt: time.Time{},
			Hashtags:  []string{},
			HasMedia:  true,
			Raw:       &bsky.FeedPost{},
		},
		{
			URI:        nsfwLabelledPost,
			ActorDID:   furry.DID(),
			CreatedAt:  time.Time{},
			IndexedAt:  time.Time{},
			Hashtags:   []string{"art"},
			HasMedia:   true,
			Raw:        &bsky.FeedPost{},
			SelfLabels: []string{"sexual"},
		},
		{
			URI:        pinnedPost,
			ActorDID:   pinnedFurry.DID(),
			CreatedAt:  time.Time{},
			IndexedAt:  time.Time{},
			Hashtags:   []string{},
			HasMedia:   true,
			Raw:        &bsky.FeedPost{},
			SelfLabels: []string{},
		},
	} {
		require.NoError(t, harness.Store.CreatePost(ctx, opts))
	}

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
			name: "nsfw only art",
			opts: chronologicalGeneratorOpts{
				generatorOpts: generatorOpts{
					Hashtags: []string{"art", "furryart"},
					IsNSFW:   tristate.True,
					HasMedia: tristate.True,
				},
			},
			expectedPosts: []string{nsfwArtPost, nsfwLabelledPost},
		},
		{
			name: "pinned post",
			opts: chronologicalGeneratorOpts{
				generatorOpts: generatorOpts{
					Hashtags: []string{"placeholder"},
				},
				PinnedDIDs: []string{pinnedFurry.DID()},
			},
			expectedPosts: []string{pinnedPost},
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
}
