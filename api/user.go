package api

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
)

type UserServiceHandler struct {
	authEngine *AuthEngine
}

func (u *UserServiceHandler) GetMe(ctx context.Context, req *connect.Request[v1.GetMeRequest]) (*connect.Response[v1.GetMeResponse], error) {
	ac, err := u.authEngine.auth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	return connect.NewResponse(&v1.GetMeResponse{
		Actor: ac.Actor,
	}), nil
}

func (u *UserServiceHandler) JoinApprovalQueue(_ context.Context, _ *connect.Request[v1.JoinApprovalQueueRequest]) (*connect.Response[v1.JoinApprovalQueueResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}
