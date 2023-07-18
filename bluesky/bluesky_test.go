package bluesky_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strideynet/bsky-furry-feed/bluesky"
)

func TestParseNamespaceID(t *testing.T) {
	assert := assert.New(t)

	nsid, err := bluesky.ParseNamespaceID("at://did:plc:bv2ckchoc76yobfhkrrie4g6/app.bsky.feed.post/3jzswcmgyao2v")
	assert.Nil(err)
	assert.Equal(nsid, "app.bsky.feed.post")

	_, err = bluesky.ParseNamespaceID("at://wrong-namespace")
	assert.Equal(err, bluesky.ErrMalformedRecordUri)
}
