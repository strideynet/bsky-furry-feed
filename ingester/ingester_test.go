package ingester_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/ingester"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"github.com/strideynet/bsky-furry-feed/testenv"
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

	now := time.Now().UTC().Truncate(time.Millisecond)
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
				Hashtags: []string{},
				HasMedia: pgtype.Bool{
					Bool:  false,
					Valid: true,
				},
				SelfLabels: []string{},
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
				SelfLabels: []string{},
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
				Hashtags: []string{
					"seni", "ısırır", "isirır",
				},
				HasMedia: pgtype.Bool{
					Bool:  false,
					Valid: true,
				},
				SelfLabels: []string{},
			},
		},
		{
			name: "hashtags in alt",
			user: approvedFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "some very undescriptive text",
				Embed: &bsky.FeedPost_Embed{
					EmbedImages: &bsky.EmbedImages{
						LexiconTypeID: "app.bsky.embed.images",
						Images: []*bsky.EmbedImages_Image{
							{
								Alt: "#fursuit #murrsuit #furryart #commsopen #nsfw #bigBurgers",
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
				SelfLabels: []string{},
			},
		},
		{
			name: "self labels",
			user: approvedFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "paws paws paws",
				Labels: &bsky.FeedPost_Labels{
					LabelDefs_SelfLabels: &atproto.LabelDefs_SelfLabels{
						Values: []*atproto.LabelDefs_SelfLabel{
							{
								Val: "adult",
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
				Hashtags: []string{},
				HasMedia: pgtype.Bool{
					Bool:  false,
					Valid: true,
				},
				SelfLabels: []string{},
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

	t.Run("waiting for posts", func(t *testing.T) {
		for _, tp := range testPosts {
			if tp.wantPost == nil {
				// Skip posts we don't expect to show up.
				continue
			}
			tp := tp
			t.Run(tp.name, func(t *testing.T) {
				t.Parallel()
				require.EventuallyWithT(t, func(t *assert.CollectT) {
					out, err := harness.Store.GetPostByURI(ctx, tp.uri)
					if !assert.NoError(t, err) {
						return
					}

					// We don't know these values at the time of initializing the test case
					// so we can set them here before assertion.
					tp.wantPost.URI = tp.uri
					tp.wantPost.Raw = tp.post
					assert.Empty(
						t,
						cmp.Diff(
							*tp.wantPost,
							out,
							// We can't know IndexedAt ahead of time.
							cmpopts.IgnoreFields(gen.CandidatePost{}, "IndexedAt"),
							cmpopts.SortSlices(func(a, b string) bool { return a < b }),
						),
					)
				}, time.Second*5, time.Millisecond*100)
			})
		}
	})

	// Now we can ensure the posts that were ignored don't show
	// TODO: We still can't be totally sure these have been ingested...
	// We need some way of telling that there's nothing left on the firehose
	// to slorp.
	for _, tp := range testPosts {
		if tp.wantPost != nil {
			continue
		}
		_, err := harness.Store.GetPostByURI(ctx, tp.uri)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	}

	// Ensure ingester closes properly
	fiCancel()
	select {
	case <-time.After(time.Second * 5):
		require.FailNow(t, "firehose ingester did not finish within deadline")
	case <-fiWait:
	}
}
