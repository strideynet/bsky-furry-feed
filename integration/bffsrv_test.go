package integration_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/api"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/feed"
	"github.com/strideynet/bsky-furry-feed/proto/bff/v1/bffv1pbconnect"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/strideynet/bsky-furry-feed/ingester"
	"github.com/strideynet/bsky-furry-feed/integration"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
)

func TestIngester(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := integration.StartHarness(ctx, t)

	bob := harness.PDS.MustNewUser(t, "bob.tpds")
	furry := harness.PDS.MustNewUser(t, "furry.tpds")

	cac := ingester.NewActorCache(harness.Log, harness.Store)
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status:  bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Comment: "furry.tpds",
		DID:     furry.DID(),
	})
	require.NoError(t, err)
	require.NoError(t, cac.Sync(ctx))

	fi := ingester.NewFirehoseIngester(harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost())
	ended := false
	defer func() { ended = true }()
	go func() {
		err := fi.Start(ctx)
		if !ended {
			require.NoError(t, err)
		}
	}()

	ignoredPost := bob.Post(t, "lorem ipsum dolor sit amet")
	trackedPost := furry.Post(t, "thank u bites u")

	// ensure ingester has processed posts
	var postURIs []string
	require.Eventually(t, func() bool {
		rows, err := harness.DBConn.Query(ctx, "select uri from candidate_posts")
		require.NoError(t, err)
		postURIs, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
			err = row.Scan(&s)
			return
		})
		require.NoError(t, err)
		return len(postURIs) > 0
	}, time.Second, 10*time.Millisecond)

	assert.Equal(t, 1, len(postURIs))
	assert.Contains(t, postURIs, trackedPost.Uri)
	assert.NotContains(t, postURIs, ignoredPost.Uri)
}

func postWithLangs(t *testing.T, u *indigoTest.TestUser, body string, langs []string) *atproto.RepoStrongRef {
	t.Helper()

	ctx := context.TODO()
	resp, err := atproto.RepoCreateRecord(ctx, integration.ExtractClientFromTestUser(u), &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       u.DID(),
		Record: &lexutil.LexiconTypeDecoder{
			Val: &bsky.FeedPost{
				CreatedAt: time.Now().Format(time.RFC3339),
				Text:      body,
				Langs:     langs,
			},
		},
	})
	require.NoError(t, err)

	return &atproto.RepoStrongRef{
		Cid: resp.Cid,
		Uri: resp.Uri,
	}
}

func TestIngester_Hashtags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := integration.StartHarness(ctx, t)

	furry := harness.PDS.MustNewUser(t, "furry.tpds")

	cac := ingester.NewActorCache(harness.Log, harness.Store)
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status:  bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Comment: "furry.tpds",
		DID:     furry.DID(),
	})
	require.NoError(t, err)
	require.NoError(t, cac.Sync(ctx))

	fi := ingester.NewFirehoseIngester(harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost())
	ended := false
	defer func() { ended = true }()
	go func() {
		err := fi.Start(ctx)
		if !ended {
			require.NoError(t, err)
		}
	}()

	enPost := postWithLangs(t, furry, "#Thank #u #bItes #U", []string{"en"})
	jaPost := postWithLangs(t, furry, "＃ありがとう ＃噛む", []string{"ja"})
	trPost := postWithLangs(t, furry, "#SENİ #ISIRır", []string{"tr"})

	// ensure ingester has processed posts
	type postAndHashtag struct {
		uri      string
		hashtags []string
	}
	var postAndHashtags []postAndHashtag
	require.Eventually(t, func() bool {
		rows, err := harness.DBConn.Query(ctx, "select uri, hashtags from candidate_posts")
		require.NoError(t, err)
		postAndHashtags, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (s postAndHashtag, err error) {
			err = row.Scan(&s.uri, &s.hashtags)
			return
		})
		require.NoError(t, err)
		return len(postAndHashtags) == 3
	}, time.Second, 10*time.Millisecond)

	for _, post := range postAndHashtags {
		var expectedHashtags []string
		switch post.uri {
		case enPost.Uri:
			expectedHashtags = []string{"thank", "bites", "u"}
		case jaPost.Uri:
			expectedHashtags = []string{"ありがとう", "噛む"}
		case trPost.Uri:
			expectedHashtags = []string{"seni", "ısırır", "isirır"}
		}
		require.ElementsMatch(t, post.hashtags, expectedHashtags)
	}
}

func TestAPI_CreateActor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := integration.StartHarness(ctx, t)

	furryActor := harness.PDS.MustNewUser(t, "furry.tpds")
	modActor := harness.PDS.MustNewUser(t, "mod.tpds")
	_ = harness.PDS.MustNewUser(t, "bff.tpds")

	srv, err := api.New(
		harness.Log,
		"",
		"",
		&feed.Service{},
		harness.Store,
		&bluesky.Credentials{
			Identifier: "bff.tpds",
			Password:   "password",
		},
		&api.AuthEngine{
			PDSHost: harness.PDS.HTTPHost(),
			ModeratorDIDs: []string{
				modActor.DID(),
			},
		},
	)
	require.NoError(t, err)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer lis.Close()
	go func() {
		err := srv.Serve(lis)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			require.NoError(t, err)
		}
	}()
	defer srv.Close()

	modActorToken := integration.ExtractClientFromTestUser(modActor).Auth.AccessJwt
	modSvc := bffv1pbconnect.NewModerationServiceClient(
		http.DefaultClient,
		"http://"+lis.Addr().String(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+modActorToken)
				return next(ctx, req)
			}
		})),
	)

	_, err = modSvc.CreateActor(ctx, connect.NewRequest(&bffv1pb.CreateActorRequest{
		ActorDid: furryActor.DID(),
		Reason:   "im testing",
	}))
	require.NoError(t, err)

	res, err := modSvc.GetActor(ctx, connect.NewRequest(&bffv1pb.GetActorRequest{
		Did: furryActor.DID(),
	}))
	require.NoError(t, err)
	require.Equal(t, furryActor.DID(), res.Msg.Actor.Did)
	require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_NONE, res.Msg.Actor.Status)
}
