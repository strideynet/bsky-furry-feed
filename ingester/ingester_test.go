package ingester_test

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/parallel"
	"github.com/bluesky-social/indigo/lex/util"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/bluesky-social/jetstream/pkg/consumer"
	jetstreamsrv "github.com/bluesky-social/jetstream/pkg/server"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
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

	cac := ingester.NewActorCache(slog.Default(), harness.Store)
	require.NoError(t, cac.Sync(ctx))

	jetstream, err := jetstreamsrv.NewServer(1)
	require.NoError(t, err)
	dataDir := t.TempDir()

	streamConsumer, err := consumer.NewConsumer(ctx, slog.Default(), "ws://"+harness.PDS.RawHost()+"/xrpc/com.atproto.sync.subscribeRepos", dataDir, time.Hour, jetstream.Emit)
	require.NoError(t, err)
	defer streamConsumer.Shutdown()
	jetstream.Consumer = streamConsumer
	streamEcho := echo.New()
	streamEcho.GET("/subscribe", jetstream.HandleSubscribe)

	go func() {
		err := streamEcho.Start(":")
		require.NoError(t, err)
	}()

	err = streamConsumer.RunSequencer(ctx)
	require.NoError(t, err)

	scheduler := parallel.NewScheduler(1, 100, "prod-firehose", streamConsumer.HandleStreamEvent)

	wsConn, _, err := websocket.DefaultDialer.Dial(streamConsumer.SocketURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { wsConn.Close() })

	go func() {
		err := events.HandleRepoStream(ctx, wsConn, scheduler, slog.Default())
		if err != nil && !strings.Contains(err.Error(), net.ErrClosed.Error()) {
			require.NoError(t, err)
		}
	}()

	fi := ingester.NewFirehoseIngester(
		slog.Default(), harness.Store, cac, "ws://"+streamEcho.Listener.Addr().String()+"/subscribe",
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
				HasVideo: pgtype.Bool{
					Bool:  false,
					Valid: true,
				},
				SelfLabels: []string{},
			},
		},
		{
			name: "video and hashtags",
			user: approvedFurry,
			post: &bsky.FeedPost{
				LexiconTypeID: "app.bsky.feed.post",
				CreatedAt:     now.Format(time.RFC3339Nano),
				Text:          "i love to poast #fursuit #murrsuit #furryart #commsopen #nsfw #bigBurgers",
				Embed: &bsky.FeedPost_Embed{
					EmbedVideo: &bsky.EmbedVideo{
						Video: &lexutil.LexBlob{
							Size:     6_000_000,
							Ref:      util.LexLink(indigoTest.RandFakeCid()),
							MimeType: "video/mp4",
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
					Bool:  false,
					Valid: true,
				},
				HasVideo: pgtype.Bool{
					Bool:  true,
					Valid: true,
				},
				SelfLabels: []string{},
			},
		},
		{
			name: "image and hashtags",
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
				HasVideo: pgtype.Bool{
					Bool:  false,
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
				HasVideo: pgtype.Bool{
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
				HasVideo: pgtype.Bool{
					Bool:  false,
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
				HasVideo: pgtype.Bool{
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
			if tp.name == "self labels" {
				t.Skip("see https://github.com/strideynet/bsky-furry-feed/issues/149")
			}
			tp := tp
			t.Run(tp.name, func(t *testing.T) {
				// todo: figure out why we can’t use parallel here!
				// t.Parallel()
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
							// Can’t compare private fields in blob
							cmpopts.IgnoreFields(lexutil.LexBlob{}, "Ref"),
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
		require.ErrorIs(t, err, store.ErrNotFound)
	}

	// Ensure ingester closes properly
	fiCancel()
	select {
	case <-time.After(time.Second * 5):
		require.FailNow(t, "firehose ingester did not finish within deadline")
	case <-fiWait:
	}
}
