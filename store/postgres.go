package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed/store")

type PGXStore struct {
	log     *zap.Logger
	pool    *pgxpool.Pool
	queries *gen.Queries
}

func (s *PGXStore) Close() {
	s.pool.Close()
}

type PoolConnector interface {
	poolConfig(ctx context.Context) (*pgxpool.Config, error)
}

func ConnectPGXStore(ctx context.Context, log *zap.Logger, connector PoolConnector) (*PGXStore, error) {
	poolCfg, err := connector.poolConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connecting pool: %w", err)
	}

	return &PGXStore{
		log:     log,
		pool:    pool,
		queries: gen.New(),
	}, nil
}

func actorStatusFromProto(s v1.ActorStatus) (gen.ActorStatus, error) {
	switch s {
	case v1.ActorStatus_ACTOR_STATUS_PENDING:
		return gen.ActorStatusPending, nil
	case v1.ActorStatus_ACTOR_STATUS_APPROVED:
		return gen.ActorStatusApproved, nil
	case v1.ActorStatus_ACTOR_STATUS_BANNED:
		return gen.ActorStatusBanned, nil
	case v1.ActorStatus_ACTOR_STATUS_NONE:
		return gen.ActorStatusNone, nil
	default:
		return "", fmt.Errorf("unhandled proto actor status: %s", s)
	}
}

func actorStatusToProto(s gen.ActorStatus) (v1.ActorStatus, error) {
	switch s {
	case gen.ActorStatusPending:
		return v1.ActorStatus_ACTOR_STATUS_PENDING, nil
	case gen.ActorStatusApproved:
		return v1.ActorStatus_ACTOR_STATUS_APPROVED, nil
	case gen.ActorStatusBanned:
		return v1.ActorStatus_ACTOR_STATUS_BANNED, nil
	case gen.ActorStatusNone:
		return v1.ActorStatus_ACTOR_STATUS_NONE, nil
	default:
		return v1.ActorStatus_ACTOR_STATUS_UNSPECIFIED, fmt.Errorf("unsupported actor status: %s", s)
	}
}

func actorToProto(actor gen.CandidateActor) (*v1.Actor, error) {
	status, err := actorStatusToProto(actor.Status)
	if err != nil {
		return nil, fmt.Errorf("converting status: %w", err)
	}
	return &v1.Actor{
		Did:       actor.DID,
		IsHidden:  actor.IsHidden,
		IsArtist:  actor.IsArtist,
		Comment:   actor.Comment,
		Status:    status,
		CreatedAt: timestamppb.New(actor.CreatedAt.Time),
	}, nil
}

func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

type ListActorsOpts struct {
	FilterStatus v1.ActorStatus
}

func (s *PGXStore) ListActors(ctx context.Context, opts ListActorsOpts) (out []*v1.Actor, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.list_actors")
	defer func() {
		endSpan(span, err)
	}()

	statusFilter := gen.NullActorStatus{}
	if opts.FilterStatus != v1.ActorStatus_ACTOR_STATUS_UNSPECIFIED {
		status, err := actorStatusFromProto(opts.FilterStatus)
		if err != nil {
			return nil, fmt.Errorf("converting filter_status: %w", err)
		}
		statusFilter.Valid = true
		statusFilter.ActorStatus = status
	}

	actors, err := s.queries.ListCandidateActors(ctx, s.pool, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("executing ListCandidateActors query: %w", err)
	}

	for _, a := range actors {
		convertedActor, err := actorToProto(a)
		if err != nil {
			return nil, fmt.Errorf("converting actor (%s): %w", a.DID, err)
		}
		out = append(out, convertedActor)
	}

	return out, nil
}

type CreateActorOpts struct {
	DID     string
	Comment string
	Status  v1.ActorStatus
}

func (s *PGXStore) CreateActor(ctx context.Context, opts CreateActorOpts) (out *v1.Actor, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_actor")
	defer func() {
		endSpan(span, err)
	}()

	status, err := actorStatusFromProto(opts.Status)
	if err != nil {
		return nil, fmt.Errorf("converting status: %w", err)
	}
	queryParams := gen.CreateCandidateActorParams{
		DID:     opts.DID,
		Comment: opts.Comment,
		Status:  status,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}
	created, err := s.queries.CreateCandidateActor(ctx, s.pool, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing CreateCandidateActor query: %w", err)
	}

	convertedActor, err := actorToProto(created)
	if err != nil {
		return nil, fmt.Errorf("converting actor (%s): %w", convertedActor.Did, err)
	}

	return convertedActor, nil
}

type UpdateActorOpts struct {
	// DID is the DID of the actor to update.
	DID string
	// Predicate is a function which is called on the fetched actor before
	// updating it. This allows rules to be placed that prevent invalid state
	// transitions.
	Predicate func(actor *v1.Actor) error
	// TODO: These fields should be optional
	UpdateStatus   v1.ActorStatus
	UpdateIsArtist bool
	UpdateComment  string
}

func (s *PGXStore) UpdateActor(ctx context.Context, opts UpdateActorOpts) (out *v1.Actor, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.update_actor")
	defer func() {
		endSpan(span, err)
	}()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			s.log.Warn("failed to roll back transaction", zap.Error(err))
		}
	}()

	dbActor, err := s.queries.GetCandidateActorByDID(ctx, tx, opts.DID)
	if err != nil {
		return nil, fmt.Errorf("fetching actor: %w", err)
	}

	actor, err := actorToProto(dbActor)
	if err != nil {
		return nil, fmt.Errorf("converting actor: %w", err)
	}
	err = opts.Predicate(actor)
	if err != nil {
		return nil, fmt.Errorf("update predicate: %w", err)
	}

	status, err := actorStatusFromProto(opts.UpdateStatus)
	if err != nil {
		return nil, fmt.Errorf("converting status: %w", err)
	}
	queryParams := gen.UpdateCandidateActorParams{
		DID: opts.DID,
		Status: gen.NullActorStatus{
			ActorStatus: status,
			Valid:       true,
		},
		IsArtist: pgtype.Bool{
			Bool:  opts.UpdateIsArtist,
			Valid: true,
		},
		Comment: pgtype.Text{
			Valid:  true,
			String: opts.UpdateComment,
		},
	}
	created, err := s.queries.UpdateCandidateActor(ctx, tx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing UpdateCandidateActor query: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	actor, err = actorToProto(created)
	if err != nil {
		return nil, fmt.Errorf("converting actor: %w", err)
	}
	return actor, nil
}

type CreateLikeOpts struct {
	URI        string
	ActorDID   string
	SubjectURI string
	CreatedAt  time.Time
	IndexedAt  time.Time
}

func (s *PGXStore) CreateLike(ctx context.Context, opts CreateLikeOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_like")
	defer func() {
		endSpan(span, err)
	}()

	queryParams := gen.CreateCandidateLikeParams{
		URI:        opts.URI,
		ActorDID:   opts.ActorDID,
		SubjectURI: opts.SubjectURI,
		CreatedAt: pgtype.Timestamptz{
			Time:  opts.CreatedAt,
			Valid: true,
		},
		IndexedAt: pgtype.Timestamptz{
			Time:  opts.IndexedAt,
			Valid: true,
		},
	}
	err = s.queries.CreateCandidateLike(ctx, s.pool, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateCandidateLike query: %w", err)
	}

	return nil
}

type DeleteLikeOpts struct {
	URI string
}

func (s *PGXStore) DeleteLike(ctx context.Context, opts DeleteLikeOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.delete_like")
	defer func() {
		endSpan(span, err)
	}()

	err = s.queries.SoftDeleteCandidateLike(ctx, s.pool, opts.URI)
	if err != nil {
		return fmt.Errorf("executing SoftDeleteCandidateLike query: %w", err)
	}

	return nil
}

type CreatePostOpts struct {
	URI       string
	ActorDID  string
	CreatedAt time.Time
	IndexedAt time.Time
	Tags      []string
}

func (s *PGXStore) CreatePost(ctx context.Context, opts CreatePostOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_post")
	defer func() {
		endSpan(span, err)
	}()

	queryParams := gen.CreateCandidatePostParams{
		URI:      opts.URI,
		ActorDID: opts.ActorDID,
		CreatedAt: pgtype.Timestamptz{
			Time:  opts.CreatedAt,
			Valid: true,
		},
		IndexedAt: pgtype.Timestamptz{
			Time:  opts.IndexedAt,
			Valid: true,
		},
		Tags: opts.Tags,
	}
	err = s.queries.CreateCandidatePost(ctx, s.pool, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateCandidatePost query: %w", err)
	}

	return nil
}

type DeletePostOpts struct {
	URI string
}

func (s *PGXStore) DeletePost(ctx context.Context, opts DeletePostOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.delete_post")
	defer func() {
		endSpan(span, err)
	}()

	err = s.queries.SoftDeleteCandidatePost(ctx, s.pool, opts.URI)
	if err != nil {
		return fmt.Errorf("executing SoftDeleteCandidatePost query: %w", err)
	}

	return nil
}

type CreateFollowOpts struct {
	URI        string
	ActorDID   string
	SubjectDID string
	CreatedAt  time.Time
	IndexedAt  time.Time
}

func (s *PGXStore) CreateFollow(ctx context.Context, opts CreateFollowOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_follow")
	defer func() {
		endSpan(span, err)
	}()

	queryParams := gen.CreateCandidateFollowParams{
		URI:        opts.URI,
		ActorDID:   opts.ActorDID,
		SubjectDid: opts.SubjectDID,
		CreatedAt: pgtype.Timestamptz{
			Time:  opts.CreatedAt,
			Valid: true,
		},
		IndexedAt: pgtype.Timestamptz{
			Time:  opts.IndexedAt,
			Valid: true,
		},
	}
	err = s.queries.CreateCandidateFollow(ctx, s.pool, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateCandidateFollowParams query: %w", err)
	}

	return nil
}

type DeleteFollowOpts struct {
	URI string
}

func (s *PGXStore) DeleteFollow(ctx context.Context, opts DeleteFollowOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.delete_follow")
	defer func() {
		endSpan(span, err)
	}()

	err = s.queries.SoftDeleteCandidateFollow(ctx, s.pool, opts.URI)
	if err != nil {
		return fmt.Errorf("executing SoftDeleteCandidateFollow query: %w", err)
	}

	return nil
}

type ListPostsForNewFeedOpts struct {
	CursorTime *time.Time
	FilterTag  string
	Limit      int
}

func (s *PGXStore) ListPostsForNewFeed(ctx context.Context, opts ListPostsForNewFeedOpts) (out []gen.CandidatePost, err error) {
	// TODO: Don't leak gen.CandidatePost implementation
	ctx, span := tracer.Start(ctx, "pgx_store.list_posts_for_new_feed")
	defer func() {
		endSpan(span, err)
	}()

	queryParams := gen.GetFurryNewFeedParams{}
	if opts.CursorTime != nil {
		queryParams.CursorTimestamp = pgtype.Timestamptz{
			Valid: true,
			Time:  *opts.CursorTime,
		}
	}
	if opts.FilterTag != "" {
		queryParams.Tag = opts.FilterTag
	}
	if opts.Limit != 0 {
		queryParams.Limit = int32(opts.Limit)
	}

	posts, err := s.queries.GetFurryNewFeed(ctx, s.pool, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing GetFurryNewFeed query: %w", err)
	}

	return posts, nil
}

type ListPostsWithLikesOpts struct {
	CursorTime *time.Time
	Limit      int
}

func (s *PGXStore) ListPostsWithLikes(ctx context.Context, opts ListPostsWithLikesOpts) (out []gen.GetPostsWithLikesRow, err error) {
	// TODO: Don't leak gen.GetPostsWithLikesRow implementation
	ctx, span := tracer.Start(ctx, "pgx_store.list_posts_with_likes")
	defer func() {
		endSpan(span, err)
	}()

	queryParams := gen.GetPostsWithLikesParams{}
	if opts.CursorTime != nil {
		queryParams.CursorTimestamp = pgtype.Timestamptz{
			Valid: true,
			Time:  *opts.CursorTime,
		}
	}
	if opts.Limit != 0 {
		queryParams.Limit = int32(opts.Limit)
	}

	posts, err := s.queries.GetPostsWithLikes(ctx, s.pool, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing GetFurryNewFeed query: %w", err)
	}

	return posts, nil
}
