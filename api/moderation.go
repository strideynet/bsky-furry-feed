package api

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"time"
)

type ModerationServiceHandler struct {
	store       *store.PGXStore
	log         *zap.Logger
	clientCache *cachedBlueSkyClient
}

func (m *ModerationServiceHandler) CreateCommentAuditEvent(ctx context.Context, req *connect.Request[v1.CreateCommentAuditEventRequest]) (*connect.Response[v1.CreateCommentAuditEventResponse], error) {
	authCtx, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	switch {
	case req.Msg.Comment == "":
		return nil, fmt.Errorf("comment is required")
	case req.Msg.SubjectDid == "":
		return nil, fmt.Errorf("subject_did is required")
	}

	ae, err := m.store.CreateAuditEvent(ctx, store.CreateAuditEventOpts{
		Payload: &v1.CommentAuditPayload{
			Comment: req.Msg.Comment,
		},
		ActorDID:   authCtx.DID,
		SubjectDID: req.Msg.SubjectDid,
	})
	if err != nil {
		return nil, fmt.Errorf("creating audit event: %w", err)
	}

	return connect.NewResponse(&v1.CreateCommentAuditEventResponse{
		AuditEvent: ae,
	}), nil
}

func (m *ModerationServiceHandler) ListAuditEvents(ctx context.Context, req *connect.Request[v1.ListAuditEventsRequest]) (*connect.Response[v1.ListAuditEventsResponse], error) {
	_, err := auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	switch {
	case req.Msg.ActorDid != "":
		return nil, fmt.Errorf("filtering by actor_did is not implemented")
	case req.Msg.SubjectRecordUri != "":
		return nil, fmt.Errorf("filtering by subject_record_uri is not implemented")
	}

	out, err := m.store.ListAuditEvents(ctx, store.ListAuditEventsOpts{
		FilterSubjectDID: req.Msg.SubjectDid,
	})
	if err != nil {
		return nil, fmt.Errorf("listing audit events: %w", err)
	}

	return connect.NewResponse(&v1.ListAuditEventsResponse{
		AuditEvents: out,
	}), nil
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

	actors, err := m.store.ListActors(ctx, store.ListActorsOpts{
		FilterStatus: req.Msg.FilterStatus,
	})
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

	actor, err := m.store.GetActorByDID(ctx, req.Msg.Did)
	if err != nil {
		return nil, fmt.Errorf("getting actor: %w", err)
	}

	res := connect.NewResponse(&v1.GetActorResponse{
		Actor: actor,
	})
	return res, nil
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
		statusToSet = v1.ActorStatus_ACTOR_STATUS_NONE
	default:
		return nil, fmt.Errorf("unsupported 'action': %s", req.Msg.Action)
	}
	actorDID := req.Msg.Did
	if actorDID == "" {
		return nil, fmt.Errorf("validating did: missing")
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
	})
	if err != nil {
		return nil, fmt.Errorf("updating actor: %w", err)
	}

	// Follow them if its an approval
	if statusToSet == v1.ActorStatus_ACTOR_STATUS_APPROVED {
		c, err := m.clientCache.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting bsky client: %w", err)
		}
		if err := c.Follow(ctx, actorDID); err != nil {
			return nil, fmt.Errorf("following approved actor: %w", err)
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err := m.store.CreateAuditEvent(ctx, store.CreateAuditEventOpts{
			Payload: &v1.ProcessApprovalQueueAuditPayload{
				Action: req.Msg.Action,
			},
			ActorDID:   authCtx.DID,
			SubjectDID: actorDID, // actor here is subject
		})
		if err != nil {
			m.log.Error("failed to emit audit event", zap.Error(err))
		}
	}()

	return connect.NewResponse(&v1.ProcessApprovalQueueResponse{}), nil
}
