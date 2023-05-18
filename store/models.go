// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package store

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CandidateActor struct {
	DID       string
	CreatedAt pgtype.Timestamptz
	IsArtist  bool
	Comment   string
}

type CandidateFollow struct {
	URI        string
	ActorDID   string
	SubjectDid string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
}

type CandidateLike struct {
	URI        string
	ActorDID   string
	SubjectURI string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
}

type CandidatePost struct {
	URI       string
	ActorDID  string
	CreatedAt pgtype.Timestamptz
	IndexedAt pgtype.Timestamptz
}
