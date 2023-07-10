package api

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

type ModerationServiceHandler struct {
	store              *store.PGXStore
	log                *zap.Logger
	blueskyCredentials *bluesky.Credentials
}

func (m *ModerationServiceHandler) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	authCtx, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	// Temporary log message - useful for debugging.
	m.log.Info("received authenticated ping!", zap.String("did", authCtx.DID))

	return connect.NewResponse(&v1.PingResponse{}), nil
}

func (m *ModerationServiceHandler) ListActors(ctx context.Context, req *connect.Request[v1.ListActorsRequest]) (*connect.Response[v1.ListActorsResponse], error) {
	_, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	actors, err := m.store.ListActors(ctx, store.ListActorsOpts{})
	if err != nil {
		return nil, fmt.Errorf("listing actors: %w", err)
	}

	res := connect.NewResponse(&v1.ListActorsResponse{
		Actors: actors,
	})
	return res, nil
}

func (m *ModerationServiceHandler) GetActor(ctx context.Context, req *connect.Request[v1.GetActorRequest]) (*connect.Response[v1.GetActorResponse], error) {
	_, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	//TODO implement me
	return nil, fmt.Errorf("unimplemented")
}

func (m *ModerationServiceHandler) GetApprovalQueue(ctx context.Context, req *connect.Request[v1.GetApprovalQueueRequest]) (*connect.Response[v1.GetApprovalQueueResponse], error) {
	_, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	actors, err := m.store.ListActors(ctx, store.ListActorsOpts{
		FilterStatus: v1.ActorStatus_ACTOR_STATUS_PENDING,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pending candidate actors: %w", err)
	}

	res := &v1.GetApprovalQueueResponse{}
	res.QueueEntriesRemaining = int32(len(actors))
	if len(actors) > 0 {
		res.QueueEntry = actors[0]
	}

	return connect.NewResponse(res), nil
}

func (m *ModerationServiceHandler) ProcessApprovalQueue(ctx context.Context, req *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error) {
	authCtx, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	// Validate request fields
	var statusToSet v1.ActorStatus
	switch req.Msg.Action {
	case v1.ApprovalQueueAction_APPROVAL_QUEUE_ACTION_APPROVE:
		statusToSet = v1.ActorStatus_ACTOR_STATUS_APPROVED
	case v1.ApprovalQueueAction_APPROVAL_QUEUE_ACTION_REJECT:
		statusToSet = v1.ActorStatus_ACTOR_STATUS_UNSPECIFIED
	default:
		return nil, fmt.Errorf("unsupported 'action': %s", req.Msg.Action)
	}
	actorDID := req.Msg.Did
	if actorDID == "" {
		return nil, fmt.Errorf("validating 'did': missing")
	}
	isArtist := req.Msg.IsArtist

	_, err = m.store.UpdateActor(ctx, store.UpdateActorOpts{
		DID: actorDID,
		Predicate: func(actor *v1.Actor) error {
			if actor.Status != v1.ActorStatus_ACTOR_STATUS_PENDING {
				return fmt.Errorf("candidate actor status was %q not %q", actor.Status, v1.ActorStatus_ACTOR_STATUS_PENDING)
			}
			return nil
		},
		UpdateStatus:   statusToSet,
		UpdateIsArtist: isArtist,
		UpdateComment:  fmt.Sprintf("Set to %s by %s using ProcessApprovalQueue", statusToSet.String(), authCtx.DID),
	})
	if err != nil {
		return nil, fmt.Errorf("updating actor: %w", err)
	}

	// Follow them if its an approval
	if statusToSet == v1.ActorStatus_ACTOR_STATUS_APPROVED {
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
