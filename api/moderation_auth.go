package api

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"golang.org/x/exp/slices"
	"strings"
)

// TODO: Add Roles to Candidate Actor schema (or a seperate schema for actual
// feed users)
var moderatorDIDs = []string{
	// Noah
	"did:plc:dllwm3fafh66ktjofzxhylwk",
	// Newton
	"did:plc:ouytv644apqbu2pm7fnp7qrj",
	// Kio
	"did:plc:o74zbazekchwk2v4twee4ekb",
	// Kev
	"did:plc:bv2ckchoc76yobfhkrrie4g6",
}

type authContext struct {
	DID string
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
	_, tokenDID, err := bluesky.ClientFromToken(ctx, authParts[1])
	if err != nil {
		return nil, fmt.Errorf("verifying token: %w", err)
	}
	if !slices.Contains(moderatorDIDs, tokenDID) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("did not associated with moderator role: %s", tokenDID))
	}

	return &authContext{
		DID: tokenDID,
	}, nil
}
