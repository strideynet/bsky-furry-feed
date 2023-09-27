package ingester

import (
	"testing"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/stretchr/testify/require"
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
		{
			name: "tags out of line",
			post: &bsky.FeedPost{
				Text:  "word",
				Tags:  []string{"foo", "FOO", "BAR"},
				Langs: nil,
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "tags in facets",
			post: &bsky.FeedPost{
				Text: "word",
				Facets: []*bsky.RichtextFacet{
					{
						Index: &bsky.RichtextFacet_ByteSlice{ByteStart: 0, ByteEnd: 2},
						Features: []*bsky.RichtextFacet_Features_Elem{
							{RichtextFacet_Tag: &bsky.RichtextFacet_Tag{LexiconTypeID: "app.bsky.richtext.facet#tag", Tag: "foo"}},
						},
					},
					{
						Index: &bsky.RichtextFacet_ByteSlice{ByteStart: 2, ByteEnd: 4},
						Features: []*bsky.RichtextFacet_Features_Elem{
							{RichtextFacet_Tag: &bsky.RichtextFacet_Tag{LexiconTypeID: "app.bsky.richtext.facet#tag", Tag: "bar"}},
						},
					},
				},
				Langs: nil,
			},
			want: []string{"foo", "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ElementsMatch(t, tt.want, normalizeHashtags(extractHashtags(tt.post), tt.post.Langs))
		})
	}
}
