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
	authEngine  *AuthEngine
}

func (m *ModerationServiceHandler) BanActor(ctx context.Context, req *connect.Request[v1.BanActorRequest]) (*connect.Response[v1.BanActorResponse], error) {
	authCtx, err := m.authEngine.auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	switch {
	case req.Msg.ActorDid == "":
		return nil, fmt.Errorf("actor_did is required")
	case req.Msg.Reason == "":
		return nil, fmt.Errorf("reason is required")
	}

	c, err := m.clientCache.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting bsky client: %w", err)
	}

	actor, err := m.store.UpdateActor(ctx, store.UpdateActorOpts{
		DID:          req.Msg.ActorDid,
		UpdateStatus: v1.ActorStatus_ACTOR_STATUS_BANNED,
	})
	if err != nil {
		return nil, fmt.Errorf("updating actor: %w", err)
	}

	go m.emitAudit(store.CreateAuditEventOpts{
		Payload: &v1.BanActorAuditPayload{
			Reason: req.Msg.Reason,
		},
		ActorDID:   authCtx.DID,
		SubjectDID: req.Msg.ActorDid,
	})

	if err := c.Unfollow(ctx, req.Msg.ActorDid); err != nil {
		return nil, fmt.Errorf("unfollowing actor: %w", err)
	}

	return connect.NewResponse(&v1.BanActorResponse{
		Actor: actor,
	}), nil
}

func (m *ModerationServiceHandler) UnapproveActor(ctx context.Context, req *connect.Request[v1.UnapproveActorRequest]) (*connect.Response[v1.UnapproveActorResponse], error) {
	authCtx, err := m.authEngine.auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	switch {
	case req.Msg.ActorDid == "":
		return nil, fmt.Errorf("actor_did is required")
	case req.Msg.Reason == "":
		return nil, fmt.Errorf("reason is required")
	}

	c, err := m.clientCache.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting bsky client: %w", err)
	}

	actor, err := m.store.UpdateActor(ctx, store.UpdateActorOpts{
		DID: req.Msg.ActorDid,
		Predicate: func(actor *v1.Actor) error {
			if actor.Status != v1.ActorStatus_ACTOR_STATUS_APPROVED {
				return fmt.Errorf("candidate actor status was %q not %q", actor.Status, v1.ActorStatus_ACTOR_STATUS_APPROVED)
			}
			return nil
		},
		UpdateStatus: v1.ActorStatus_ACTOR_STATUS_NONE,
	})
	if err != nil {
		return nil, fmt.Errorf("updating actor: %w", err)
	}

	go m.emitAudit(store.CreateAuditEventOpts{
		Payload: &v1.UnapproveActorAuditPayload{
			Reason: req.Msg.Reason,
		},
		ActorDID:   authCtx.DID,
		SubjectDID: req.Msg.ActorDid,
	})

	if err := c.Unfollow(ctx, req.Msg.ActorDid); err != nil {
		return nil, fmt.Errorf("unfollowing actor: %w", err)
	}

	return connect.NewResponse(&v1.UnapproveActorResponse{
		Actor: actor,
	}), nil
}

func (m *ModerationServiceHandler) CreateActor(ctx context.Context, req *connect.Request[v1.CreateActorRequest]) (*connect.Response[v1.CreateActorResponse], error) {
	authCtx, err := m.authEngine.auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	switch {
	case req.Msg.ActorDid == "":
		return nil, fmt.Errorf("actor_did is required")
	case req.Msg.Reason == "":
		return nil, fmt.Errorf("reason is required")
	}

	actor, err := m.store.CreateActor(ctx, store.CreateActorOpts{
		Status:  v1.ActorStatus_ACTOR_STATUS_NONE,
		DID:     req.Msg.ActorDid,
		Comment: "",
	})
	if err != nil {
		return nil, fmt.Errorf("creating actor: %w", err)
	}

	go m.emitAudit(store.CreateAuditEventOpts{
		Payload: &v1.CreateActorAuditPayload{
			Reason: req.Msg.Reason,
		},
		ActorDID:   authCtx.DID,
		SubjectDID: req.Msg.ActorDid,
	})

	return connect.NewResponse(&v1.CreateActorResponse{
		Actor: actor,
	}), nil
}

func (m *ModerationServiceHandler) CreateCommentAuditEvent(ctx context.Context, req *connect.Request[v1.CreateCommentAuditEventRequest]) (*connect.Response[v1.CreateCommentAuditEventResponse], error) {
	authCtx, err := m.authEngine.auth(ctx, req)
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
		ActorDID:         authCtx.DID,
		SubjectDID:       req.Msg.SubjectDid,
		SubjectRecordURI: req.Msg.SubjectRecordUri,
	})
	if err != nil {
		return nil, fmt.Errorf("creating audit event: %w", err)
	}

	return connect.NewResponse(&v1.CreateCommentAuditEventResponse{
		AuditEvent: ae,
	}), nil
}

func (m *ModerationServiceHandler) ListAuditEvents(ctx context.Context, req *connect.Request[v1.ListAuditEventsRequest]) (*connect.Response[v1.ListAuditEventsResponse], error) {
	_, err := m.authEngine.auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	switch {
	case req.Msg.FilterActorDid != "":
		return nil, fmt.Errorf("filter_actor_did is not implemented")
	case req.Msg.FilterSubjectRecordUri != "":
		return nil, fmt.Errorf("filter_subject_record_uri is not implemented")
	}

	out, err := m.store.ListAuditEvents(ctx, store.ListAuditEventsOpts{
		FilterSubjectDID: req.Msg.FilterSubjectDid,
	})
	if err != nil {
		return nil, fmt.Errorf("listing audit events: %w", err)
	}

	return connect.NewResponse(&v1.ListAuditEventsResponse{
		AuditEvents: out,
	}), nil
}

func (m *ModerationServiceHandler) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	authCtx, err := m.authEngine.auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	// Temporary log message - useful for debugging.
	m.log.Info("received authenticated ping!", zap.String("did", authCtx.DID))

	return connect.NewResponse(&v1.PingResponse{}), nil
}

func (m *ModerationServiceHandler) ListActors(ctx context.Context, req *connect.Request[v1.ListActorsRequest]) (*connect.Response[v1.ListActorsResponse], error) {
	_, err := m.authEngine.auth(ctx, req)
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
	_, err := m.authEngine.auth(ctx, req)
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

func (m *ModerationServiceHandler) ProcessApprovalQueue(ctx context.Context, req *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error) {
	authCtx, err := m.authEngine.auth(ctx, req)
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

	c, err := m.clientCache.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting bsky client: %w", err)
	}

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
		if err := c.Follow(ctx, actorDID); err != nil {
			return nil, fmt.Errorf("following approved actor: %w", err)
		}
	}

	go m.emitAudit(store.CreateAuditEventOpts{
		Payload: &v1.ProcessApprovalQueueAuditPayload{
			Action: req.Msg.Action,
		},
		ActorDID:   authCtx.DID,
		SubjectDID: actorDID,
	})

	return connect.NewResponse(&v1.ProcessApprovalQueueResponse{}), nil
}

func (m *ModerationServiceHandler) emitAudit(opts store.CreateAuditEventOpts) {
	// TODO: Consider pulling this out of a goroutine and making it part
	// of the transaction in the database?
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := m.store.CreateAuditEvent(ctx, opts)
	if err != nil {
		m.log.Error("failed to emit audit event", zap.Error(err))
	}
}
