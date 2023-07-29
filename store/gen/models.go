// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package gen

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type ActorStatus string

const (
	ActorStatusNone     ActorStatus = "none"
	ActorStatusPending  ActorStatus = "pending"
	ActorStatusApproved ActorStatus = "approved"
	ActorStatusBanned   ActorStatus = "banned"
)

func (e *ActorStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ActorStatus(s)
	case string:
		*e = ActorStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ActorStatus: %T", src)
	}
	return nil
}

type NullActorStatus struct {
	ActorStatus ActorStatus
	Valid       bool // Valid is true if ActorStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullActorStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ActorStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ActorStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullActorStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ActorStatus), nil
}

type AuditEvent struct {
	ID               string
	ActorDID         string
	SubjectDid       string
	SubjectRecordUri string
	CreatedAt        pgtype.Timestamptz
	Payload          []byte
}

type CandidateActor struct {
	DID       string
	CreatedAt pgtype.Timestamptz
	IsArtist  bool
	Comment   string
	IsNSFW    bool
	IsHidden  bool
	Status    ActorStatus
}

type CandidateFollow struct {
	URI        string
	ActorDID   string
	SubjectDid string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

type CandidateLike struct {
	URI        string
	ActorDID   string
	SubjectURI string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

type CandidatePost struct {
	URI       string
	ActorDID  string
	CreatedAt pgtype.Timestamptz
	IndexedAt pgtype.Timestamptz
	IsNSFW    bool
	IsHidden  bool
	Tags      []string
	DeletedAt pgtype.Timestamptz
}
