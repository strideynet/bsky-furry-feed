package api

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestAPI_WellKnownDID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx := context.Background()
	harness := startAPIHarness(ctx, t)
	resp, err := http.Get(harness.APIAddr + "/.well-known/did.json")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	want := `{"@context":["https://www.w3.org/ns/did/v1"],"id":"did:web:feed.test.furryli.st","service":[{"id":"#bsky_fg","type":"BskyFeedGenerator","serviceEndpoint":"https://feed.test.furryli.st"}]}` + "\n"
	require.Equal(t, want, string(bytes))
}
