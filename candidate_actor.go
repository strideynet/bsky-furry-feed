package bff

import (
	"github.com/strideynet/bsky-furry-feed/store"
	"time"
)

type CandidateActor struct {
	DID       string    `json:"did"`
	CreatedAt time.Time `json:"created_at"`
	IsArtist  bool      `json:"is_artist"`
	Comment   string    `json:"comment"`
}

func CandidateActorFromStore(cr store.CandidateActor) CandidateActor {
	return CandidateActor{
		DID:       cr.DID,
		CreatedAt: cr.CreatedAt.Time,
		IsArtist:  cr.IsArtist,
		Comment:   cr.Comment,
	}
}
