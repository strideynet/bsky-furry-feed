package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/strideynet/bsky-furry-feed/bfflog"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
)

type actorGetter interface {
	GetActorByDID(ctx context.Context, did string) (*v1.Actor, error)
}

func BSkyTokenValidator(pdsHost string) func(ctx context.Context, token string) (did string, err error) {
	// Check the presented token is valid against the real bsky.
	// This also lets us introspect information about the user - we can't just
	// parse the JWT as they do not use public key signing for the JWT.
	return func(ctx context.Context, token string) (did string, err error) {
		ua := bluesky.UserAgent
		res, err := getBlueskySession(ctx, &xrpc.Client{
			Host:      pdsHost,
			UserAgent: &ua,
			Auth:      &xrpc.AuthInfo{AccessJwt: token},
		})
		if err != nil {
			return "", fmt.Errorf("get session: %w", err)
		}
		return res.Did, nil
	}
}

type ServerGetSession_Output struct {
	Did            string  `json:"did" cborgen:"did"`
	Email          *string `json:"email,omitempty" cborgen:"email,omitempty"`
	EmailConfirmed *bool   `json:"emailConfirmed,omitempty" cborgen:"emailConfirmed,omitempty"`
	Handle         string  `json:"handle" cborgen:"handle"`
}

// Workaround until Bluesky’s util.LexiconTypeDecoder for
// DidDoc doesn’t always error with `unrecognized type: ""`
func getBlueskySession(ctx context.Context, c *xrpc.Client) (*ServerGetSession_Output, error) {
	var out ServerGetSession_Output
	if err := c.Do(ctx, xrpc.Query, "", "com.atproto.server.getSession", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// authenticatedUserPermissions are granted to any user who is authenticated.
var authenticatedUserPermissions = []string{
	"/bff.v1.ModerationService/Ping",
	"/bff.v1.UserService/GetMe",
	"/bff.v1.UserService/JoinApprovalQueue",
}

var approverPermissions = []string{
	"/bff.v1.ModerationService/GetActor",
	"/bff.v1.ModerationService/ListActors",
	"/bff.v1.ModerationService/ListAuditEvents",
	"/bff.v1.ModerationService/ProcessApprovalQueue",
	"/bff.v1.ModerationService/CreateCommentAuditEvent",
	"/bff.v1.ModerationService/HoldBackPendingActor",
	"/bff.v1.ModerationService/ListRoles",
}

var moderatorPermissions = append([]string{
	"/bff.v1.ModerationService/UnapproveActor",
	"/bff.v1.ModerationService/ForceApproveActor",
}, approverPermissions...)

var adminPermissions = append([]string{
	"/bff.v1.ModerationService/BanActor",
	"/bff.v1.ModerationService/CreateActor",
	"/bff.v1.ModerationService/AssignRoles",
}, moderatorPermissions...)

var roleToPermissions = map[string][]string{
	"admin":     adminPermissions,
	"moderator": moderatorPermissions,
	"approver":  approverPermissions,
}

// AuthEngine helps authenticate requests made by users and apply authorization
// rules based on the identity found during authentication.
type AuthEngine struct {
	// ActorGetter provides a way for the AuthEngine to fetch the Actor data
	// associated with a given DID.
	ActorGetter actorGetter
	// TokenValidator validates a given token and returns the DID associated
	// with that token.
	TokenValidator func(ctx context.Context, token string) (did string, err error)
	Log            *slog.Logger
}

type authContext struct {
	// DID is the did extracted from the token supplied by the user.
	DID string
	// Actor is the actor fetched from the database during authz/authn. This
	// should be used carefully, and if necessary the actor should be fetched
	// again within a transaction if mutation is occurring.
	//
	// This will be nil if the actor does not exist.
	Actor *v1.Actor
}

// TODO: Allow a authOpts to be passed in with a description of attempted
// action.
func (a *AuthEngine) auth(ctx context.Context, req connect.AnyRequest) (*authContext, error) {
	// Extract the token from the headers
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

	// Validate the token from the header
	did, err := a.TokenValidator(ctx, authParts[1])
	if err != nil {
		return nil, fmt.Errorf("validating token: %w", err)
	}

	// Try to fetch the actor to find any roles they have associated with them.
	// If they don't exist - we continue - so act with caution, actor may be
	// nil.
	actor, err := a.ActorGetter.GetActorByDID(ctx, did)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("fetching actor for token: %w", err)
	}

	// Default to no roles if the actor does not exist.
	var actorRoles []string
	if actor != nil {
		actorRoles = actor.Roles
	}

	// Convert the actors roles to a quasi-set of permitted RPCs
	permissions := map[string]bool{}
	for _, permission := range authenticatedUserPermissions {
		permissions[permission] = true
	}
	for _, role := range actorRoles {
		rolePerms, ok := roleToPermissions[role]
		if !ok {
			// Gracefully handle an unrecognized role
			a.Log.Warn(
				"unrecognized role",
				slog.String("role", role),
				bfflog.ActorDID(did),
			)
			continue
		}
		for _, permission := range rolePerms {
			permissions[permission] = true
		}
	}

	// Check user has permission for target RPC
	procedureName := req.Spec().Procedure
	if !permissions[procedureName] {
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			fmt.Errorf("user (%s) does not have permissions for %q", did, procedureName),
		)
	}

	return &authContext{
		DID:   did,
		Actor: actor,
	}, nil
}
