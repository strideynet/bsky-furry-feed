package api

import (
	"context"
	"github.com/bufbuild/connect-go"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/moderation/v1"
)

type ModerationServiceHandler struct {
}

func (m *ModerationServiceHandler) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return connect.NewResponse(&v1.PingResponse{}), nil
}

func (m *ModerationServiceHandler) GetApprovalQueue(ctx context.Context, req *connect.Request[v1.GetApprovalQueueRequest]) (*connect.Response[v1.GetApprovalQueueResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (m *ModerationServiceHandler) ProcessApprovalQueue(ctx context.Context, req *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error) {
	//TODO implement me
	panic("implement me")
}
