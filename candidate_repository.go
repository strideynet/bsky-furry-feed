package bff

import (
	"github.com/strideynet/bsky-furry-feed/store"
	"time"
)

type CandidateRepository struct {
	DID       string
	CreatedAt time.Time
	IsArtist  bool
	Comment   string
}

func CandidateRepositoryFromStore(cr store.CandidateRepository) CandidateRepository {
	return CandidateRepository{
		DID:       cr.DID,
		CreatedAt: cr.CreatedAt.Time,
		IsArtist:  cr.IsArtist,
		Comment:   cr.Comment,
	}
}
