package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/xid"
	"github.com/strideynet/bsky-furry-feed/bfflog"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"github.com/strideynet/bsky-furry-feed/tristate"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed/store")

type PGXStore struct {
	log     *slog.Logger
	pool    *pgxpool.Pool
	queries *gen.Queries
}

func (s *PGXStore) Close() {
	s.pool.Close()
}

func convertPGXError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	default:
		return err
	}
}

type PoolConnector interface {
	poolConfig(ctx context.Context) (*pgxpool.Config, error)
}

func ConnectPGXStore(ctx context.Context, log *slog.Logger, connector PoolConnector) (*PGXStore, error) {
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
		queries: gen.New(pool),
	}, nil
}

// RawPool returns the underlying [pgxpool.Pool] used by the store. This should
// be avoided in production.
func (s *PGXStore) RawPool() *pgxpool.Pool {
	return s.pool
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
		IsArtist:  actor.IsArtist,
		Comment:   actor.Comment,
		Status:    status,
		CreatedAt: timestamppb.New(actor.CreatedAt.Time),
		Roles:     actor.Roles,
		HeldUntil: timestamppb.New(actor.HeldUntil.Time),
	}, nil
}

func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

type PGXTX struct {
	*PGXStore
	tx pgx.Tx
}

// Rollback rolls back a transaction if it has not already been committed.
// This is safe to call when the transaction has been committed, so defer this
// when creating a transaction.
func (s *PGXTX) Rollback() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := s.tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		s.PGXStore.log.Error(
			"failed to rollback transaction", bfflog.Err(err),
		)
	}
}

func (s *PGXTX) Commit(ctx context.Context) error {
	// return without wrapping since it'll just result in duplication of
	// "committing transaction".
	return s.tx.Commit(ctx)
}

func (s *PGXTX) TX(_ context.Context) (*PGXStore, error) {
	// TODO(noah): Evaluate if we can support nested TX.
	return nil, fmt.Errorf("nested transactions not supported")
}

func (s *PGXStore) TX(ctx context.Context) (*PGXTX, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}

	return &PGXTX{
		tx: tx,
		PGXStore: &PGXStore{
			log:     s.log,
			pool:    s.pool,
			queries: gen.New(tx),
		},
	}, nil
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

	actors, err := s.queries.ListCandidateActors(ctx, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("executing ListCandidateActors query: %w", convertPGXError(err))
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

func (s *PGXStore) GetActorByDID(ctx context.Context, did string) (out *v1.Actor, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.get_actor_by_did")
	defer func() {
		endSpan(span, err)
	}()

	actor, err := s.queries.GetCandidateActorByDID(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("executing GetCandidateActorByDID query: %w", convertPGXError(err))
	}

	out, err = actorToProto(actor)
	if err != nil {
		return nil, fmt.Errorf("converting actor (%s): %w", actor.DID, err)
	}

	return out, nil
}

type CreateActorOpts struct {
	DID     string
	Comment string
	Status  v1.ActorStatus
	Roles   []string
}

func (s *PGXStore) CreateActor(ctx context.Context, opts CreateActorOpts) (out *v1.Actor, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_actor")
	defer func() {
		endSpan(span, err)
	}()

	if opts.Roles == nil {
		opts.Roles = []string{}
	}

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
		Roles: opts.Roles,
	}
	created, err := s.queries.CreateCandidateActor(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing CreateCandidateActor query: %w", convertPGXError(err))
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
	// TODO: These fields should be optional
	UpdateStatus   v1.ActorStatus
	UpdateIsArtist bool
	UpdateComment  string
	UpdateRoles    []string
}

func (s *PGXStore) UpdateActor(ctx context.Context, opts UpdateActorOpts) (out *v1.Actor, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.update_actor")
	defer func() {
		endSpan(span, err)
	}()

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
		Roles: opts.UpdateRoles,
	}
	created, err := s.queries.UpdateCandidateActor(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing UpdateCandidateActor query: %w", convertPGXError(err))
	}

	actor, err := actorToProto(created)
	if err != nil {
		return nil, fmt.Errorf("converting actor: %w", err)
	}
	return actor, nil
}

type CreateLatestActorProfileOpts struct {
	// DID is the DID of the actor to update.
	ActorDID string
	// CommitCID is now used for the repo rev at the commit
	// that set this version of the profile.
	CommitCID   string
	CreatedAt   time.Time
	IndexedAt   time.Time
	DisplayName string
	Description string
	SelfLabels  []string
}

func (s *PGXStore) CreateLatestActorProfile(ctx context.Context, opts CreateLatestActorProfileOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.update_actor_profile")
	defer func() {
		endSpan(span, err)
	}()

	if opts.SelfLabels == nil {
		opts.SelfLabels = []string{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			s.log.Warn("failed to roll back transaction", bfflog.Err(err))
		}
	}()

	queryParams := gen.CreateLatestActorProfileParams{
		ActorDID:  opts.ActorDID,
		CommitCID: opts.CommitCID,
		CreatedAt: pgtype.Timestamptz{
			Valid: true,
			Time:  opts.CreatedAt,
		},
		IndexedAt: pgtype.Timestamptz{
			Time:  opts.IndexedAt,
			Valid: true,
		},
		DisplayName: pgtype.Text{
			Valid:  true,
			String: opts.DisplayName,
		},
		Description: pgtype.Text{
			Valid:  true,
			String: opts.Description,
		},
		SelfLabels: opts.SelfLabels,
	}
	err = s.queries.CreateLatestActorProfile(ctx, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateLatestActorProfile query: %w", convertPGXError(err))
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
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
	err = s.queries.CreateCandidateLike(ctx, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateCandidateLike query: %w", convertPGXError(err))
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

	err = s.queries.SoftDeleteCandidateLike(ctx, opts.URI)
	if err != nil {
		return fmt.Errorf("executing SoftDeleteCandidateLike query: %w", convertPGXError(err))
	}

	return nil
}

type CreatePostOpts struct {
	URI        string
	ActorDID   string
	CreatedAt  time.Time
	IndexedAt  time.Time
	Hashtags   []string
	HasMedia   bool
	HasVideo   bool
	Raw        *bsky.FeedPost
	SelfLabels []string
}

func (s *PGXStore) CreatePost(ctx context.Context, opts CreatePostOpts) (err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_post")
	defer func() {
		endSpan(span, err)
	}()

	if opts.SelfLabels == nil {
		opts.SelfLabels = []string{}
	}

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
		Hashtags: opts.Hashtags,
		HasMedia: pgtype.Bool{
			Valid: true,
			Bool:  opts.HasMedia,
		},
		HasVideo: pgtype.Bool{
			Valid: true,
			Bool:  opts.HasVideo,
		},
		Raw:        opts.Raw,
		SelfLabels: opts.SelfLabels,
	}
	err = s.queries.CreateCandidatePost(ctx, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateCandidatePost query: %w", convertPGXError(err))
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

	err = s.queries.SoftDeleteCandidatePost(ctx, opts.URI)
	if err != nil {
		return fmt.Errorf("executing SoftDeleteCandidatePost query: %w", convertPGXError(err))
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
	err = s.queries.CreateCandidateFollow(ctx, queryParams)
	if err != nil {
		return fmt.Errorf("executing CreateCandidateFollowParams query: %w", convertPGXError(err))
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

	err = s.queries.SoftDeleteCandidateFollow(ctx, opts.URI)
	if err != nil {
		return fmt.Errorf("executing SoftDeleteCandidateFollow query: %w", convertPGXError(err))
	}

	return nil
}

type ListPostsForNewFeedOpts struct {
	CursorTime         time.Time
	Hashtags           []string
	DisallowedHashtags []string
	IsNSFW             tristate.Tristate
	AllowedEmbeds      []string
	PinnedDIDs         []string
	Limit              int
}

func tristateToPgtypeBool(t tristate.Tristate) pgtype.Bool {
	b := (*bool)(t)
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Valid: true, Bool: *b}
}

func (s *PGXStore) ListPostsForNewFeed(ctx context.Context, opts ListPostsForNewFeedOpts) (out []gen.CandidatePost, err error) {
	// TODO: Don't leak gen.CandidatePost implementation
	ctx, span := tracer.Start(ctx, "pgx_store.list_posts_for_new_feed")
	defer func() {
		endSpan(span, err)
	}()
	queryParams := gen.GetFurryNewFeedParams{
		CursorTimestamp: pgtype.Timestamptz{
			Valid: true,
			Time:  opts.CursorTime,
		},
		Hashtags:           opts.Hashtags,
		DisallowedHashtags: opts.DisallowedHashtags,
		AllowedEmbeds:      opts.AllowedEmbeds,
		IsNSFW:             tristateToPgtypeBool(opts.IsNSFW),
		PinnedDIDs:         opts.PinnedDIDs,
	}
	if opts.Limit != 0 {
		queryParams.Limit = int32(opts.Limit)
	}

	posts, err := s.queries.GetFurryNewFeed(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing GetFurryNewFeed query: %w", convertPGXError(err))
	}

	return posts, nil
}

func (s *PGXStore) GetLatestScoreGeneration(ctx context.Context, alg string) (out int64, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.get_latest_score_generation")
	defer func() {
		endSpan(span, err)
	}()
	seq, err := s.queries.GetLatestScoreGeneration(ctx, alg)
	if err != nil {
		return 0, convertPGXError(err)
	}
	return seq, nil
}

type ListPostsForHotFeedCursor struct {
	GenerationSeq int64
	AfterScore    float32
	AfterURI      string
}

type ListPostsForHotFeedOpts struct {
	Alg                string
	Cursor             ListPostsForHotFeedCursor
	Hashtags           []string
	DisallowedHashtags []string
	IsNSFW             tristate.Tristate
	AllowedEmbeds      []string
	Limit              int
}

func (s *PGXStore) ListScoredPosts(ctx context.Context, opts ListPostsForHotFeedOpts) (out []gen.ListScoredPostsRow, err error) {
	// TODO: Don't leak gen.CandidatePost implementation
	ctx, span := tracer.Start(ctx, "pgx_store.list_scored_posts")
	defer func() {
		endSpan(span, err)
	}()

	queryParams := gen.ListScoredPostsParams{
		Alg:                opts.Alg,
		Hashtags:           opts.Hashtags,
		DisallowedHashtags: opts.DisallowedHashtags,
		AllowedEmbeds:      opts.AllowedEmbeds,
		IsNSFW:             tristateToPgtypeBool(opts.IsNSFW),
		GenerationSeq:      opts.Cursor.GenerationSeq,
		AfterScore:         opts.Cursor.AfterScore,
		AfterURI:           opts.Cursor.AfterURI,
	}
	if opts.Limit != 0 {
		queryParams.Limit = int32(opts.Limit)
	}

	posts, err := s.queries.ListScoredPosts(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing ListScoredPosts query: %w", convertPGXError(err))
	}

	return posts, nil
}

type ListPostsWithLikesOpts struct {
	CursorTime time.Time
	Limit      int
}

func auditEventToProto(in gen.AuditEvent) (*v1.AuditEvent, error) {
	ae := &v1.AuditEvent{
		Id:               in.ID,
		CreatedAt:        timestamppb.New(in.CreatedAt.Time),
		ActorDid:         in.ActorDID,
		SubjectDid:       in.SubjectDid,
		SubjectRecordUri: in.SubjectRecordUri,
	}
	anyPayload := &anypb.Any{}
	err := protojson.Unmarshal(in.Payload, anyPayload)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling payload: %w", err)
	}
	ae.Payload = anyPayload
	return ae, nil
}

type ListAuditEventsOpts struct {
	FilterSubjectDID    string
	FilterCreatedBefore *time.Time
	FilterTypes         []v1.AuditEventType

	// Limit defaults to 100.
	Limit int32
}

func (s *PGXStore) ListAuditEvents(ctx context.Context, opts ListAuditEventsOpts) (out []*v1.AuditEvent, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.list_audit_events")
	defer func() {
		endSpan(span, err)
	}()

	limit := opts.Limit
	if limit == 0 {
		limit = 100
	}

	queryParams := gen.ListAuditEventsParams{
		SubjectDid: opts.FilterSubjectDID,
		Limit:      limit,
	}
	if opts.FilterCreatedBefore != nil {
		queryParams.CreatedBefore = pgtype.Timestamptz{
			Time:  *opts.FilterCreatedBefore,
			Valid: true,
		}
	}
	for _, typ := range opts.FilterTypes {
		queryParams.Types = append(queryParams.Types, typ.String())
	}

	data, err := s.queries.ListAuditEvents(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing ListAuditEvents query: %w", convertPGXError(err))
	}

	out = make([]*v1.AuditEvent, 0, len(data))
	for _, d := range data {
		ae, err := auditEventToProto(d)
		if err != nil {
			return nil, fmt.Errorf("converting audit event: %w", err)
		}
		out = append(out, ae)
	}

	return out, nil
}

type CreateAuditEventOpts struct {
	ActorDID         string
	SubjectDID       string
	SubjectRecordURI string
	Payload          proto.Message
}

func (s *PGXStore) CreateAuditEvent(ctx context.Context, opts CreateAuditEventOpts) (out *v1.AuditEvent, err error) {
	ctx, span := tracer.Start(ctx, "pgx_store.create_audit_event")
	defer func() {
		endSpan(span, err)
	}()

	payload, err := anypb.New(opts.Payload)
	if err != nil {
		return nil, fmt.Errorf("creating anypb: %w", err)
	}
	payloadBytes, err := protojson.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling anypb: %w", err)
	}
	queryParams := gen.CreateAuditEventParams{
		ID: xid.New().String(),
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ActorDID:         opts.ActorDID,
		SubjectDid:       opts.SubjectDID,
		SubjectRecordUri: opts.SubjectRecordURI,
		Payload:          payloadBytes,
	}
	data, err := s.queries.CreateAuditEvent(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("executing CreateAuditEvent query: %w", convertPGXError(err))
	}

	out, err = auditEventToProto(data)
	if err != nil {
		return nil, fmt.Errorf("converting inserted audit event to proto: %w", err)
	}

	return out, nil
}

func (s *PGXStore) GetJetstreamCursor(ctx context.Context) (out int64, err error) {
	out, err = s.queries.GetJetstreamCursor(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Special sentinel value for no cursor persisted.
			return -1, nil
		}
		return 0, err
	}
	return out, nil
}

func (s *PGXStore) SetJetstreamCursor(ctx context.Context, cursor int64) (err error) {
	return convertPGXError(s.queries.SetJetstreamCursor(ctx, cursor))
}

func (s *PGXStore) GetPostByURI(ctx context.Context, uri string) (out gen.CandidatePost, err error) {
	// TODO: Return a proto type rather than exposing gen.CandidatePost
	out, err = s.queries.GetPostByURI(ctx, uri)
	return out, convertPGXError(err)
}

func (s *PGXStore) GetLatestActorProfile(ctx context.Context, did string) (out gen.ActorProfile, err error) {
	// TODO: Return a proto type rather than exposing gen.ActorProfile
	out, err = s.queries.GetLatestActorProfile(ctx, did)
	return out, convertPGXError(err)
}

func (s *PGXStore) GetActorProfileHistory(ctx context.Context, did string) (out []gen.ActorProfile, err error) {
	// TODO: Return a proto type rather than exposing gen.ActorProfile
	out, err = s.queries.GetActorProfileHistory(ctx, did)
	return out, convertPGXError(err)
}

func (s *PGXStore) MaterializeClassicPostScores(ctx context.Context, after time.Time) (int64, error) {
	return s.queries.MaterializePostScores(ctx, pgtype.Timestamptz{Time: after, Valid: true})
}

func (s *PGXStore) DeleteOldPostScores(ctx context.Context, before time.Time) (int64, error) {
	return s.queries.DeleteOldPostScores(ctx, pgtype.Timestamptz{Time: before, Valid: true})
}

func (s *PGXStore) HoldBackPendingActor(ctx context.Context, did string, duration time.Time) error {
	return s.queries.HoldBackPendingActor(ctx, gen.HoldBackPendingActorParams{
		DID:       did,
		HeldUntil: pgtype.Timestamptz{Time: duration, Valid: true},
	})
}

func (s *PGXStore) EnqueueFollow(ctx context.Context, did string) error {
	return s.queries.EnqueueFollowTask(ctx, gen.EnqueueFollowTaskParams{
		ActorDID:       did,
		NextTryAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ShouldUnfollow: false,
	})
}

func (s *PGXStore) EnqueueUnfollow(ctx context.Context, did string) error {
	return s.queries.EnqueueFollowTask(ctx, gen.EnqueueFollowTaskParams{
		ActorDID:       did,
		NextTryAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ShouldUnfollow: true,
	})
}

func (s *PGXStore) GetNextFollowTask(ctx context.Context) (gen.FollowTask, error) {
	return s.queries.GetNextFollowTask(ctx)
}

func (s *PGXStore) MarkFollowTaskAsErrored(ctx context.Context, id int64, err error) error {
	return s.queries.MarkFollowTaskAsErrored(ctx, gen.MarkFollowTaskAsErroredParams{
		ID:        id,
		LastError: pgtype.Text{String: err.Error(), Valid: true},
		NextTryAt: pgtype.Timestamptz{Time: time.Now().Add(time.Minute * 15), Valid: true},
	})
}

func (s *PGXStore) MarkFollowTaskAsDone(ctx context.Context, id int64) error {
	return s.queries.MarkFollowTaskAsDone(ctx, id)
}
