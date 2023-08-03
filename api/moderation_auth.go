package api

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"strings"
)

type actorGetter interface {
	GetActorByDID(ctx context.Context, did string) (*v1.Actor, error)
}

// authenticatedUserPermissions are granted to any user who is authenticated.
var authenticatedUserPermissions = []string{
	"/bff.v1.ModerationService/Ping",
}

var moderatorPermissions = []string{
	"/bff.v1.ModerationService/GetActor",
	"/bff.v1.ModerationService/ListActors",
	"/bff.v1.ModerationService/ListAuditEvents",
	"/bff.v1.ModerationService/ProcessApprovalQueue",
	"/bff.v1.ModerationService/CreateCommentAuditEvent",
}

var adminPermissions = append([]string{
	"/bff.v1.ModerationService/BanActor",
	"/bff.v1.ModerationService/UnapproveActor",
	"/bff.v1.ModerationService/ForceApproveActor",
	"/bff.v1.ModerationService/CreateActor",
}, moderatorPermissions...)

var roleToPermissions = map[string][]string{
	"admin":     adminPermissions,
	"moderator": moderatorPermissions,
}

type AuthEngine struct {
	actorGetter   actorGetter
	ModeratorDIDs []string
	PDSHost       string
	Log           *zap.Logger
}

type authContext struct {
	DID string
}

// TODO: Allow a authOpts to be passed in with a description of attempted
// action.
func (a *AuthEngine) auth(ctx context.Context, req connect.AnyRequest) (*authContext, error) {
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
	_, tokenDID, err := bluesky.ClientFromToken(ctx, a.PDSHost, authParts[1])
	if err != nil {
		return nil, fmt.Errorf("verifying token: %w", err)
	}
	if !slices.Contains(a.ModeratorDIDs, tokenDID) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("did not associated with moderator role: %s", tokenDID))
	}

	// Calculate user permissions
	roles := []string{}
	permissions := map[string]bool{}
	// We know the user is authenticated so we grant them the authenticated
	// user role.
	for _, permission := range authenticatedUserPermissions {
		permissions[permission] = true
	}
	// Now we grant them all the permissions from their roles
	for _, role := range roles {
		rolePerms, ok := roleToPermissions[role]
		if !ok {
			a.Log.Warn("unrecognized role", zap.String("role", role), zap.String("actor_did", tokenDID))
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
			fmt.Errorf("user (%s) does not have permissions for %q", tokenDID, procedureName),
		)
	}

	return &authContext{
		DID: tokenDID,
	}, nil
}
