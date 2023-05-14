package bff

import (
	"github.com/strideynet/bsky-furry-feed/store"
	"time"
)

type CandidateRepository struct {
	DID       string    `json:"did"`
	CreatedAt time.Time `json:"created_at"`
	IsArtist  bool      `json:"is_artist"`
	Comment   string    `json:"comment"`
}

func CandidateRepositoryFromStore(cr store.CandidateRepository) CandidateRepository {
	return CandidateRepository{
		DID:       cr.DID,
		CreatedAt: cr.CreatedAt.Time,
		IsArtist:  cr.IsArtist,
		Comment:   cr.Comment,
	}
}
