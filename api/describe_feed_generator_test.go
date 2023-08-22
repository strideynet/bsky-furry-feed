package api

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestAPI_DescribeFeedGenerator(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx := context.Background()
	harness := startAPIHarness(ctx, t)
	resp, err := http.Get(harness.APIAddr + "/xrpc/app.bsky.feed.describeFeedGenerator")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	want := `{"did":"did:web:feed.test.furryli.st","feeds":[{"uri":"at://did:web:feed.test.furryli.st/app.bsky.feed.generator/fake-1"}]}` + "\n"
	require.Equal(t, want, string(bytes))
}
