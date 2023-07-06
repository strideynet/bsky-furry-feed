package api

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/moderation/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

type ModerationServiceHandler struct {
	queries            *store.QueriesWithTX
	log                *zap.Logger
	blueskyCredentials *bluesky.Credentials
}

func (m *ModerationServiceHandler) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	err := auth(ctx, req)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.PingResponse{}), nil
}

func (m *ModerationServiceHandler) GetApprovalQueue(ctx context.Context, req *connect.Request[v1.GetApprovalQueueRequest]) (*connect.Response[v1.GetApprovalQueueResponse], error) {
	err := auth(ctx, req)
	if err != nil {
		return nil, err
	}

	actors, err := m.queries.ListCandidateActors(ctx, store.NullActorStatus{
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
	err := auth(ctx, req)
	if err != nil {
		return nil, err
	}

	// Validate request fields
	var statusToSet store.ActorStatus
	switch req.Msg.Action {
	case v1.ApprovalQueueAction_APPROVAL_QUEUE_ACTION_APPROVE:
		statusToSet = store.ActorStatusApproved
	case v1.ApprovalQueueAction_APPROVAL_QUEUE_ACTION_REJECT:
		statusToSet = store.ActorStatusNone
	default:
		return nil, fmt.Errorf("unsupported 'action': %s", req.Msg.Action)
	}
	actorDID := req.Msg.Did
	if actorDID == "" {
		return nil, fmt.Errorf("validating 'did': missing")
	}
	isArtist := req.Msg.IsArtist

	// Open transaction to make sure we don't double process an actor
	queries, commit, rollback, err := m.queries.WithTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer rollback()

	// Fetch specified actor to ensure it exists and is pending
	actor, err := queries.GetCandidateActorByDID(ctx, actorDID)
	if err != nil {
		return nil, fmt.Errorf("fetching candidate actor: %w", err)
	}
	if actor.Status != store.ActorStatusPending {
		return nil, fmt.Errorf("candidate actor status was %q not %q", actor.Status, store.ActorStatusPending)
	}

	_, err = queries.UpdateCandidateActor(ctx, store.UpdateCandidateActorParams{
		DID: actorDID,
		Status: store.NullActorStatus{
			ActorStatus: statusToSet,
			Valid:       true,
		},
		IsArtist: pgtype.Bool{
			Bool:  isArtist,
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
	if err := commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	// Follow them if its an approval
	if statusToSet == store.ActorStatusApproved {
		bskyClient, err := bluesky.ClientFromCredentials(ctx, m.blueskyCredentials)
		if err != nil {
			return nil, fmt.Errorf("creating bsky client: %w", err)
		}
		if err := bskyClient.Follow(ctx, actorDID); err != nil {
			return nil, fmt.Errorf("following approved actor: %w", err)
		}
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
