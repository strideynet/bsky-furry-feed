package api

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/moderation/v1"
	"github.com/strideynet/bsky-furry-feed/store"
)

type ModerationServiceHandler struct {
	db *pgxpool.Pool
}

func (m *ModerationServiceHandler) Ping(_ context.Context, _ *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return connect.NewResponse(&v1.PingResponse{}), nil
}

func (m *ModerationServiceHandler) GetApprovalQueue(ctx context.Context, _ *connect.Request[v1.GetApprovalQueueRequest]) (*connect.Response[v1.GetApprovalQueueResponse], error) {
	queries := store.New(m.db)
	actors, err := queries.ListCandidateActors(ctx, store.NullActorStatus{
		Valid:       true,
		ActorStatus: store.ActorStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pending candidate actors: %w", err)
	}

	res := &v1.GetApprovalQueueResponse{}
	res.QueueEntriesRemaining = int32(len(actors))
	if len(actors) > 0 {
		res.QueueEntry = candidateActorToProto(actors[0])
	}

	return connect.NewResponse(res), nil
}

func (m *ModerationServiceHandler) ProcessApprovalQueue(ctx context.Context, req *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error) {
	// Validate request fields
	var status store.ActorStatus
	switch req.Msg.Action {
	case v1.ApprovalQueueAction_APPROVAL_QUEUE_ACTION_APPROVE:
		status = store.ActorStatusApproved
	case v1.ApprovalQueueAction_APPROVAL_QUEUE_ACTION_REJECT:
		status = store.ActorStatusNone
	default:
		return nil, fmt.Errorf("unsupported 'action': %s", req.Msg.Action)
	}

	if req.Msg.Did == "" {
		return nil, fmt.Errorf("validating 'did': missing")
	}

	// Open transaction to make sure we don't double process an actor
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(context.Background())
	queries := store.New(m.db).WithTx(tx)

	// Fetch specified actor to ensure it exists and is pending
	actor, err := queries.GetCandidateActorByDID(ctx, req.Msg.Did)
	if err != nil {
		return nil, fmt.Errorf("fetching candidate actor: %w", err)
	}
	if actor.Status != store.ActorStatusPending {
		return nil, fmt.Errorf("candidate actor status was %q not %q", actor.Status, store.ActorStatusPending)
	}

	_, err = queries.UpdateCandidateActor(ctx, store.UpdateCandidateActorParams{
		DID: req.Msg.Did,
		Status: store.NullActorStatus{
			ActorStatus: status,
			Valid:       true,
		},
		IsArtist: pgtype.Bool{
			Bool:  req.Msg.IsArtist,
			Valid: true,
		},
		Comment: pgtype.Text{
			Valid: true,
			// TODO: Calculate comment
			String: "",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("updating candidate actor: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	// TODO: Follow them if its an approval
	if status == store.ActorStatusApproved {
		fmt.Println("TODO: Approve Actor")
	}

	return connect.NewResponse(&v1.ProcessApprovalQueueResponse{}), nil
}

func candidateActorToProto(actor store.CandidateActor) *v1.CandidateActor {
	return &v1.CandidateActor{
		Did:      actor.DID,
		IsHidden: actor.IsHidden,
		IsArtist: actor.IsArtist,
		Comment:  actor.Comment,
	}
}
