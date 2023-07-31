package feed

import (
	"context"
	"testing"
	"time"

	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/integration"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
)

const createPostQuery = `
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at, hashtags, tags, has_media, raw)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8);
`

func TestChronologicalGenerator(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := integration.StartHarness(ctx, t)

	furry := harness.PDS.MustNewUser(t, "furry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status:  bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Comment: "furry.tpds",
		DID:     furry.DID(),
	})
	require.NoError(t, err)

	pool := harness.Store.TESTONLY_GetPool()

	artWithMediaURI := indigoTest.RandFakeAtUri("app.bsky.feed.post", "")
	_, err = pool.Exec(
		ctx,
		createPostQuery,

		artWithMediaURI, // uri
		furry.DID(),     // actor_did
		time.Now(),      // created_at
		time.Now(),      // indexed_at
		[]string{"art"}, // hashtags
		[]string{},      // tags
		true,            // has_media
		nil,             // raw
	)
	require.NoError(t, err)

	legacyArtWithMediaURI := indigoTest.RandFakeAtUri("app.bsky.feed.post", "")
	_, err = pool.Exec(
		ctx,
		createPostQuery,

		legacyArtWithMediaURI, // uri
		furry.DID(),           // actor_did
		time.Now(),            // created_at
		time.Now(),            // indexed_at
		[]string{},            // hashtags
		[]string{"art"},       // tags
		nil,                   // has_media
		nil,                   // raw
	)
	require.NoError(t, err)

	nsfwArtWithMediaURI := indigoTest.RandFakeAtUri("app.bsky.feed.post", "")
	_, err = pool.Exec(
		ctx,
		createPostQuery,

		nsfwArtWithMediaURI,     // uri
		furry.DID(),             // actor_did
		time.Now(),              // created_at
		time.Now(),              // indexed_at
		[]string{"art", "nsfw"}, // hashtags
		[]string{},              // tags
		true,                    // has_media
		nil,                     // raw
	)
	require.NoError(t, err)

	legacyNsfwArtWithMediaURI := indigoTest.RandFakeAtUri("app.bsky.feed.post", "")
	_, err = pool.Exec(
		ctx,
		createPostQuery,

		legacyNsfwArtWithMediaURI, // uri
		furry.DID(),               // actor_did
		time.Now(),                // created_at
		time.Now(),                // indexed_at
		[]string{},                // hashtags
		[]string{"art", "nsfw"},   // tags
		nil,                       // has_media
		nil,                       // raw
	)
	require.NoError(t, err)

	artWithNoMediaURI := indigoTest.RandFakeAtUri("app.bsky.feed.post", "")
	_, err = pool.Exec(
		ctx,
		createPostQuery,

		artWithNoMediaURI, // uri
		furry.DID(),       // actor_did
		time.Now(),        // created_at
		time.Now(),        // indexed_at
		[]string{"art"},   // hashtags
		[]string{},        // tags
		false,             // has_media
		nil,               // raw
	)
	require.NoError(t, err)

	posts, err := chronologicalGenerator(chronologicalGeneratorOpts{
		RequireTags: []string{"art"},
		ExcludeTags: []string{"nsfw"},

		IncludeHashtags: []string{"art"},
		ExcludeHashtags: []string{"nsfw"},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	})(ctx, harness.Store, "", 100)
	require.NoError(t, err)

	postURIs := make([]string, len(posts))
	for i, post := range posts {
		postURIs[i] = post.URI
	}

	require.ElementsMatch(t, postURIs, []string{artWithMediaURI, legacyArtWithMediaURI})
}

func TestChronologicalGenerator_IncludeHashtags(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := integration.StartHarness(ctx, t)

	furry := harness.PDS.MustNewUser(t, "furry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status:  bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Comment: "furry.tpds",
		DID:     furry.DID(),
	})
	require.NoError(t, err)

	pool := harness.Store.TESTONLY_GetPool()

	fursuitFridayPostURI := indigoTest.RandFakeAtUri("app.bsky.feed.post", "")
	_, err = pool.Exec(
		ctx,
		createPostQuery,

		fursuitFridayPostURI,      // uri
		furry.DID(),               // actor_did
		time.Now(),                // created_at
		time.Now(),                // indexed_at
		[]string{"fursuitfriday"}, // hashtags
		[]string{},                // tags
		true,                      // has_media
		nil,                       // raw
	)
	require.NoError(t, err)

	posts, err := chronologicalGenerator(chronologicalGeneratorOpts{
		IncludeHashtags: []string{"fursuit", "fursuitfriday"},
		HasMedia:        pgtype.Bool{Valid: true, Bool: true},
	})(ctx, harness.Store, "", 100)
	require.NoError(t, err)

	postURIs := make([]string, len(posts))
	for i, post := range posts {
		postURIs[i] = post.URI
	}

	require.ElementsMatch(t, postURIs, []string{fursuitFridayPostURI})
}
