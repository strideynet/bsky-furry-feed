package ingester_test

import (
	"context"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/ingester"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"github.com/strideynet/bsky-furry-feed/testenv"
	"testing"
	"time"
)

// TestFirehoseIngester intends to fully integration test the ingester against
// a real database and a stood up Go PDS firehose. As little as possible should
// be faked out.
//
// Where possible - integrate new test cases into this test - or create lower
// level unit tests.
func TestFirehoseIngester(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)

	nonFurry := harness.PDS.MustNewUser(t, "non-furry.tpds")
	pendingFurry := harness.PDS.MustNewUser(t, "pending-furry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_PENDING,
		DID:    pendingFurry.DID(),
	})
	require.NoError(t, err)
	approvedFurry := harness.PDS.MustNewUser(t, "approvedFurry.tpds")
	_, err = harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		DID:    approvedFurry.DID(),
	})
	require.NoError(t, err)

	cac := ingester.NewActorCache(harness.Log, harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := ingester.NewFirehoseIngester(
		harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)
	fiContext, fiCancel := context.WithCancel(ctx)
	fiWait := make(chan struct{})
	go func() {
		if err := fi.Start(fiContext); err != nil {
			if fiContext.Err() == nil {
				require.NoError(t, err)
			}
		}
		close(fiWait)
	}()

	now := time.Now().UTC()
	testPosts := []struct {
		name string
		user *indigoTest.TestUser
		post *bsky.FeedPost

		wantPost *gen.CandidatePost

		uri string
	}{
		{
			name: "non furry ignored",
			user: nonFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "lorem ipsum dolor sit amet",
			},
		},
		{
			name: "pending furry ignored",
			user: pendingFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "lorem ipsum dolor sit amet",
			},
		},
		{
			name: "simple furry",
			user: approvedFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "paws paws paws",
			},
			wantPost: &gen.CandidatePost{
				ActorDID: approvedFurry.DID(),
				CreatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
				Tags:     []string{},
				Hashtags: []string{},
				HasMedia: pgtype.Bool{
					Bool:  false,
					Valid: true,
				},
			},
		},
		{
			name: "media and hashtags",
			user: approvedFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "i love to poast #fursuit #murrsuit #furryart #commsopen #nsfw #bigBurgers",
				Embed: &bsky.FeedPost_Embed{
					EmbedImages: &bsky.EmbedImages{
						LexiconTypeID: "app.bsky.embed.images",
						Images: []*bsky.EmbedImages_Image{
							{
								Alt: "some alt text",
							},
						},
					},
				},
			},
			wantPost: &gen.CandidatePost{
				ActorDID: approvedFurry.DID(),
				CreatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
				Tags: []string{
					bff.TagFursuitMedia,
					bff.TagArt,
					bff.TagNSFW,
					bff.TagCommissionsOpen,
				},
				Hashtags: []string{
					"fursuit",
					"murrsuit",
					"furryart",
					"commsopen",
					"nsfw",
					"bigburgers",
				},
				HasMedia: pgtype.Bool{
					Bool:  true,
					Valid: true,
				},
			},
		},
		{
			name: "internationalised hashtags",
			user: approvedFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "#SENİ #ISIRır",
				Langs:         []string{"tr"},
			},
			wantPost: &gen.CandidatePost{
				ActorDID: approvedFurry.DID(),
				CreatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
				Tags: []string{},
				Hashtags: []string{
					"seni", "ısırır", "isirır",
				},
				HasMedia: pgtype.Bool{
					Bool:  false,
					Valid: true,
				},
			},
		},
	}

	for i, tp := range testPosts {
		resp, err := atproto.RepoCreateRecord(ctx, testenv.ExtractClientFromTestUser(tp.user), &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Repo:       tp.user.DID(),
			Record: &lexutil.LexiconTypeDecoder{
				Val: tp.post,
			},
		})
		require.NoError(t, err)

		// We don't know the URI until we post it and this makes assertion
		// more difficult. So we persist the returned URI here.
		testPosts[i].uri = resp.Uri
	}

	// TODO(noah): This sucks - let's see if we can detect that they have all
	// been delivered. Perhaps we can make use of the cursor updates and wait
	// until the cursor reaches the end ?
	time.Sleep(time.Second * 1)

	// Close down ingester - and wait for it to close *properly*
	fiCancel()
	select {
	case <-time.After(time.Second * 5):
		require.FailNow(t, "firehose ingester did not finish within deadline")
	case <-fiWait:
	}

	for _, tp := range testPosts {
		tp := tp
		t.Run("post: "+tp.name, func(t *testing.T) {
			t.Parallel()

			out, err := harness.Store.GetPostByURI(ctx, tp.uri)
			if tp.wantPost == nil {
				require.ErrorIs(t, err, pgx.ErrNoRows)
				return
			}
			require.NoError(t, err)

			// We don't know these values at the time of initializing the test case
			// so we can set them here before assertion.
			tp.wantPost.URI = tp.uri
			tp.wantPost.Raw = tp.post
			require.Empty(
				t,
				cmp.Diff(
					*tp.wantPost,
					out,
					// We can't know IndexedAt ahead of time.
					cmpopts.IgnoreFields(gen.CandidatePost{}, "IndexedAt"),
					cmpopts.SortSlices(func(a, b string) bool { return a < b }),
				),
			)
		})
	}
}
