// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: bff/v1/moderation_service.proto

package bffv1pbconnect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// ModerationServiceName is the fully-qualified name of the ModerationService service.
	ModerationServiceName = "bff.v1.ModerationService"
)

// ModerationServiceClient is a client for the bff.v1.ModerationService service.
type ModerationServiceClient interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	GetApprovalQueue(context.Context, *connect_go.Request[v1.GetApprovalQueueRequest]) (*connect_go.Response[v1.GetApprovalQueueResponse], error)
	// TODO: Refactor ProcessApprovalQueue to something more like "ApproveActor"
	ProcessApprovalQueue(context.Context, *connect_go.Request[v1.ProcessApprovalQueueRequest]) (*connect_go.Response[v1.ProcessApprovalQueueResponse], error)
	ListActors(context.Context, *connect_go.Request[v1.ListActorsRequest]) (*connect_go.Response[v1.ListActorsResponse], error)
	GetActor(context.Context, *connect_go.Request[v1.GetActorRequest]) (*connect_go.Response[v1.GetActorResponse], error)
	ListAuditEvents(context.Context, *connect_go.Request[v1.ListAuditEventsRequest]) (*connect_go.Response[v1.ListAuditEventsResponse], error)
	CreateCommentAuditEvent(context.Context, *connect_go.Request[v1.CreateCommentAuditEventRequest]) (*connect_go.Response[v1.CreateCommentAuditEventResponse], error)
}

// NewModerationServiceClient constructs a client for the bff.v1.ModerationService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewModerationServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ModerationServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &moderationServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/Ping",
			opts...,
		),
		getApprovalQueue: connect_go.NewClient[v1.GetApprovalQueueRequest, v1.GetApprovalQueueResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/GetApprovalQueue",
			opts...,
		),
		processApprovalQueue: connect_go.NewClient[v1.ProcessApprovalQueueRequest, v1.ProcessApprovalQueueResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/ProcessApprovalQueue",
			opts...,
		),
		listActors: connect_go.NewClient[v1.ListActorsRequest, v1.ListActorsResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/ListActors",
			opts...,
		),
		getActor: connect_go.NewClient[v1.GetActorRequest, v1.GetActorResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/GetActor",
			opts...,
		),
		listAuditEvents: connect_go.NewClient[v1.ListAuditEventsRequest, v1.ListAuditEventsResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/ListAuditEvents",
			opts...,
		),
		createCommentAuditEvent: connect_go.NewClient[v1.CreateCommentAuditEventRequest, v1.CreateCommentAuditEventResponse](
			httpClient,
			baseURL+"/bff.v1.ModerationService/CreateCommentAuditEvent",
			opts...,
		),
	}
}

// moderationServiceClient implements ModerationServiceClient.
type moderationServiceClient struct {
	ping                    *connect_go.Client[v1.PingRequest, v1.PingResponse]
	getApprovalQueue        *connect_go.Client[v1.GetApprovalQueueRequest, v1.GetApprovalQueueResponse]
	processApprovalQueue    *connect_go.Client[v1.ProcessApprovalQueueRequest, v1.ProcessApprovalQueueResponse]
	listActors              *connect_go.Client[v1.ListActorsRequest, v1.ListActorsResponse]
	getActor                *connect_go.Client[v1.GetActorRequest, v1.GetActorResponse]
	listAuditEvents         *connect_go.Client[v1.ListAuditEventsRequest, v1.ListAuditEventsResponse]
	createCommentAuditEvent *connect_go.Client[v1.CreateCommentAuditEventRequest, v1.CreateCommentAuditEventResponse]
}

// Ping calls bff.v1.ModerationService.Ping.
func (c *moderationServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetApprovalQueue calls bff.v1.ModerationService.GetApprovalQueue.
func (c *moderationServiceClient) GetApprovalQueue(ctx context.Context, req *connect_go.Request[v1.GetApprovalQueueRequest]) (*connect_go.Response[v1.GetApprovalQueueResponse], error) {
	return c.getApprovalQueue.CallUnary(ctx, req)
}

// ProcessApprovalQueue calls bff.v1.ModerationService.ProcessApprovalQueue.
func (c *moderationServiceClient) ProcessApprovalQueue(ctx context.Context, req *connect_go.Request[v1.ProcessApprovalQueueRequest]) (*connect_go.Response[v1.ProcessApprovalQueueResponse], error) {
	return c.processApprovalQueue.CallUnary(ctx, req)
}

// ListActors calls bff.v1.ModerationService.ListActors.
func (c *moderationServiceClient) ListActors(ctx context.Context, req *connect_go.Request[v1.ListActorsRequest]) (*connect_go.Response[v1.ListActorsResponse], error) {
	return c.listActors.CallUnary(ctx, req)
}

// GetActor calls bff.v1.ModerationService.GetActor.
func (c *moderationServiceClient) GetActor(ctx context.Context, req *connect_go.Request[v1.GetActorRequest]) (*connect_go.Response[v1.GetActorResponse], error) {
	return c.getActor.CallUnary(ctx, req)
}

// ListAuditEvents calls bff.v1.ModerationService.ListAuditEvents.
func (c *moderationServiceClient) ListAuditEvents(ctx context.Context, req *connect_go.Request[v1.ListAuditEventsRequest]) (*connect_go.Response[v1.ListAuditEventsResponse], error) {
	return c.listAuditEvents.CallUnary(ctx, req)
}

// CreateCommentAuditEvent calls bff.v1.ModerationService.CreateCommentAuditEvent.
func (c *moderationServiceClient) CreateCommentAuditEvent(ctx context.Context, req *connect_go.Request[v1.CreateCommentAuditEventRequest]) (*connect_go.Response[v1.CreateCommentAuditEventResponse], error) {
	return c.createCommentAuditEvent.CallUnary(ctx, req)
}

// ModerationServiceHandler is an implementation of the bff.v1.ModerationService service.
type ModerationServiceHandler interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	GetApprovalQueue(context.Context, *connect_go.Request[v1.GetApprovalQueueRequest]) (*connect_go.Response[v1.GetApprovalQueueResponse], error)
	// TODO: Refactor ProcessApprovalQueue to something more like "ApproveActor"
	ProcessApprovalQueue(context.Context, *connect_go.Request[v1.ProcessApprovalQueueRequest]) (*connect_go.Response[v1.ProcessApprovalQueueResponse], error)
	ListActors(context.Context, *connect_go.Request[v1.ListActorsRequest]) (*connect_go.Response[v1.ListActorsResponse], error)
	GetActor(context.Context, *connect_go.Request[v1.GetActorRequest]) (*connect_go.Response[v1.GetActorResponse], error)
	ListAuditEvents(context.Context, *connect_go.Request[v1.ListAuditEventsRequest]) (*connect_go.Response[v1.ListAuditEventsResponse], error)
	CreateCommentAuditEvent(context.Context, *connect_go.Request[v1.CreateCommentAuditEventRequest]) (*connect_go.Response[v1.CreateCommentAuditEventResponse], error)
}

// NewModerationServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewModerationServiceHandler(svc ModerationServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/bff.v1.ModerationService/Ping", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/Ping",
		svc.Ping,
		opts...,
	))
	mux.Handle("/bff.v1.ModerationService/GetApprovalQueue", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/GetApprovalQueue",
		svc.GetApprovalQueue,
		opts...,
	))
	mux.Handle("/bff.v1.ModerationService/ProcessApprovalQueue", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/ProcessApprovalQueue",
		svc.ProcessApprovalQueue,
		opts...,
	))
	mux.Handle("/bff.v1.ModerationService/ListActors", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/ListActors",
		svc.ListActors,
		opts...,
	))
	mux.Handle("/bff.v1.ModerationService/GetActor", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/GetActor",
		svc.GetActor,
		opts...,
	))
	mux.Handle("/bff.v1.ModerationService/ListAuditEvents", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/ListAuditEvents",
		svc.ListAuditEvents,
		opts...,
	))
	mux.Handle("/bff.v1.ModerationService/CreateCommentAuditEvent", connect_go.NewUnaryHandler(
		"/bff.v1.ModerationService/CreateCommentAuditEvent",
		svc.CreateCommentAuditEvent,
		opts...,
	))
	return "/bff.v1.ModerationService/", mux
}

// UnimplementedModerationServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedModerationServiceHandler struct{}

func (UnimplementedModerationServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.Ping is not implemented"))
}

func (UnimplementedModerationServiceHandler) GetApprovalQueue(context.Context, *connect_go.Request[v1.GetApprovalQueueRequest]) (*connect_go.Response[v1.GetApprovalQueueResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.GetApprovalQueue is not implemented"))
}

func (UnimplementedModerationServiceHandler) ProcessApprovalQueue(context.Context, *connect_go.Request[v1.ProcessApprovalQueueRequest]) (*connect_go.Response[v1.ProcessApprovalQueueResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.ProcessApprovalQueue is not implemented"))
}

func (UnimplementedModerationServiceHandler) ListActors(context.Context, *connect_go.Request[v1.ListActorsRequest]) (*connect_go.Response[v1.ListActorsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.ListActors is not implemented"))
}

func (UnimplementedModerationServiceHandler) GetActor(context.Context, *connect_go.Request[v1.GetActorRequest]) (*connect_go.Response[v1.GetActorResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.GetActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) ListAuditEvents(context.Context, *connect_go.Request[v1.ListAuditEventsRequest]) (*connect_go.Response[v1.ListAuditEventsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.ListAuditEvents is not implemented"))
}

func (UnimplementedModerationServiceHandler) CreateCommentAuditEvent(context.Context, *connect_go.Request[v1.CreateCommentAuditEventRequest]) (*connect_go.Response[v1.CreateCommentAuditEventResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("bff.v1.ModerationService.CreateCommentAuditEvent is not implemented"))
}
