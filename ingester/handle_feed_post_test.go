package ingester

import (
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_extractNormalizedHashtags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		post *bsky.FeedPost
		want []string
	}{
		{
			name: "no language",
			post: &bsky.FeedPost{
				Text:  "word #FOO #foo word2 #BAr",
				Langs: nil,
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "en",
			post: &bsky.FeedPost{
				Text:  "word #FOO #foo word2 #BAr",
				Langs: []string{"en"},
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "ja",
			post: &bsky.FeedPost{
				Text:  "＃ありがとう ＃噛む",
				Langs: []string{"ja"},
			},
			want: []string{"ありがとう", "噛む"},
		},
		{
			name: "tr",
			post: &bsky.FeedPost{
				Text:  "#SENİ #ISIRır",
				Langs: []string{"tr"},
			},
			want: []string{"seni", "ısırır", "isirır"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ElementsMatch(t, tt.want, extractNormalizedHashtags(tt.post))
		})
	}
}
