package api

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/moderation/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"strings"
)

type ModerationServiceHandler struct {
	queries *store.QueriesWithTX
	log     *zap.Logger
}

func (m *ModerationServiceHandler) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	_, err := auth(ctx, req)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.PingResponse{}), nil
}

func (m *ModerationServiceHandler) GetApprovalQueue(ctx context.Context, _ *connect.Request[v1.GetApprovalQueueRequest]) (*connect.Response[v1.GetApprovalQueueResponse], error) {
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
	queries, commit, rollback, err := m.queries.WithTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer rollback()

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
	if err := commit(ctx); err != nil {
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

type authContext struct {
	DID string
}

// TODO: Pull these from the database
var moderatorDIDs = []string{
	"did:plc:dllwm3fafh66ktjofzxhylwk",
}

// TODO: Allow a authOpts to be passed in with a description of attempted
// action.
func auth(ctx context.Context, req connect.AnyRequest) (*authContext, error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("no token provided"))
	}

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("malformed header"))
	}

	if authParts[0] != "Bearer" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("only Bearer auth supported"))
	}

	// Check the presented token is valid against the real bsky.
	// This also lets us introspect information about the user - we can't just
	// parse the JWT as they do not use public key signing for the JWT.
	bskyClient := bluesky.NewClient(&xrpc.AuthInfo{
		AccessJwt: authParts[1],
	})
	session, err := bskyClient.GetSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching session with token: %w", err)
	}

	if !slices.Contains(moderatorDIDs, session.Did) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("did not associated with moderator role: %s", session.Did))
	}

	return &authContext{DID: session.Did}, nil
}
