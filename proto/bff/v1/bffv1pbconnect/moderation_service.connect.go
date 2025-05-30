// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: bff/v1/moderation_service.proto

package bffv1pbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion0_1_0

const (
	// ModerationServiceName is the fully-qualified name of the ModerationService service.
	ModerationServiceName = "bff.v1.ModerationService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ModerationServicePingProcedure is the fully-qualified name of the ModerationService's Ping RPC.
	ModerationServicePingProcedure = "/bff.v1.ModerationService/Ping"
	// ModerationServiceProcessApprovalQueueProcedure is the fully-qualified name of the
	// ModerationService's ProcessApprovalQueue RPC.
	ModerationServiceProcessApprovalQueueProcedure = "/bff.v1.ModerationService/ProcessApprovalQueue"
	// ModerationServiceHoldBackPendingActorProcedure is the fully-qualified name of the
	// ModerationService's HoldBackPendingActor RPC.
	ModerationServiceHoldBackPendingActorProcedure = "/bff.v1.ModerationService/HoldBackPendingActor"
	// ModerationServiceListActorsProcedure is the fully-qualified name of the ModerationService's
	// ListActors RPC.
	ModerationServiceListActorsProcedure = "/bff.v1.ModerationService/ListActors"
	// ModerationServiceGetActorProcedure is the fully-qualified name of the ModerationService's
	// GetActor RPC.
	ModerationServiceGetActorProcedure = "/bff.v1.ModerationService/GetActor"
	// ModerationServiceBanActorProcedure is the fully-qualified name of the ModerationService's
	// BanActor RPC.
	ModerationServiceBanActorProcedure = "/bff.v1.ModerationService/BanActor"
	// ModerationServiceUnapproveActorProcedure is the fully-qualified name of the ModerationService's
	// UnapproveActor RPC.
	ModerationServiceUnapproveActorProcedure = "/bff.v1.ModerationService/UnapproveActor"
	// ModerationServiceForceApproveActorProcedure is the fully-qualified name of the
	// ModerationService's ForceApproveActor RPC.
	ModerationServiceForceApproveActorProcedure = "/bff.v1.ModerationService/ForceApproveActor"
	// ModerationServiceCreateActorProcedure is the fully-qualified name of the ModerationService's
	// CreateActor RPC.
	ModerationServiceCreateActorProcedure = "/bff.v1.ModerationService/CreateActor"
	// ModerationServiceListAuditEventsProcedure is the fully-qualified name of the ModerationService's
	// ListAuditEvents RPC.
	ModerationServiceListAuditEventsProcedure = "/bff.v1.ModerationService/ListAuditEvents"
	// ModerationServiceCreateCommentAuditEventProcedure is the fully-qualified name of the
	// ModerationService's CreateCommentAuditEvent RPC.
	ModerationServiceCreateCommentAuditEventProcedure = "/bff.v1.ModerationService/CreateCommentAuditEvent"
	// ModerationServiceListRolesProcedure is the fully-qualified name of the ModerationService's
	// ListRoles RPC.
	ModerationServiceListRolesProcedure = "/bff.v1.ModerationService/ListRoles"
	// ModerationServiceAssignRolesProcedure is the fully-qualified name of the ModerationService's
	// AssignRoles RPC.
	ModerationServiceAssignRolesProcedure = "/bff.v1.ModerationService/AssignRoles"
)

// ModerationServiceClient is a client for the bff.v1.ModerationService service.
type ModerationServiceClient interface {
	// Ping is a test RPC that checks that the user is authenticated and then
	// returns an empty response. Ideal for health checking the moderation service.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// TODO: Refactor ProcessApprovalQueue to something more like "ProcessPendingActor"
	ProcessApprovalQueue(context.Context, *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error)
	// HoldBackPendingActor ignores a pending actor for review in some time, so we
	// don’t need to reject actors that e.g. have no avatar or bio yet.
	HoldBackPendingActor(context.Context, *connect.Request[v1.HoldBackPendingActorRequest]) (*connect.Response[v1.HoldBackPendingActorResponse], error)
	// ListActors fetches multiple actors from the database. It allows this to be
	// filtered by certain attributes.
	ListActors(context.Context, *connect.Request[v1.ListActorsRequest]) (*connect.Response[v1.ListActorsResponse], error)
	// GetActor fetches a single actor from the database.
	GetActor(context.Context, *connect.Request[v1.GetActorRequest]) (*connect.Response[v1.GetActorResponse], error)
	// BanActor changes an actors status to "banned".
	// Actor can be in any status before they are banned.
	BanActor(context.Context, *connect.Request[v1.BanActorRequest]) (*connect.Response[v1.BanActorResponse], error)
	// UnapproveActor changes an actor from "approved" status to "none" status.
	UnapproveActor(context.Context, *connect.Request[v1.UnapproveActorRequest]) (*connect.Response[v1.UnapproveActorResponse], error)
	// ForceApproveActor changes an actor to "approved" status from "none" or "pending".
	ForceApproveActor(context.Context, *connect.Request[v1.ForceApproveActorRequest]) (*connect.Response[v1.ForceApproveActorResponse], error)
	// CreateActor creates a database entry for an actor who does not currently exist.
	// By default, their status will be set to none.
	CreateActor(context.Context, *connect.Request[v1.CreateActorRequest]) (*connect.Response[v1.CreateActorResponse], error)
	ListAuditEvents(context.Context, *connect.Request[v1.ListAuditEventsRequest]) (*connect.Response[v1.ListAuditEventsResponse], error)
	CreateCommentAuditEvent(context.Context, *connect.Request[v1.CreateCommentAuditEventRequest]) (*connect.Response[v1.CreateCommentAuditEventResponse], error)
	ListRoles(context.Context, *connect.Request[v1.ListRolesRequest]) (*connect.Response[v1.ListRolesResponse], error)
	AssignRoles(context.Context, *connect.Request[v1.AssignRolesRequest]) (*connect.Response[v1.AssignRolesResponse], error)
}

// NewModerationServiceClient constructs a client for the bff.v1.ModerationService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewModerationServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ModerationServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &moderationServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+ModerationServicePingProcedure,
			opts...,
		),
		processApprovalQueue: connect.NewClient[v1.ProcessApprovalQueueRequest, v1.ProcessApprovalQueueResponse](
			httpClient,
			baseURL+ModerationServiceProcessApprovalQueueProcedure,
			opts...,
		),
		holdBackPendingActor: connect.NewClient[v1.HoldBackPendingActorRequest, v1.HoldBackPendingActorResponse](
			httpClient,
			baseURL+ModerationServiceHoldBackPendingActorProcedure,
			opts...,
		),
		listActors: connect.NewClient[v1.ListActorsRequest, v1.ListActorsResponse](
			httpClient,
			baseURL+ModerationServiceListActorsProcedure,
			opts...,
		),
		getActor: connect.NewClient[v1.GetActorRequest, v1.GetActorResponse](
			httpClient,
			baseURL+ModerationServiceGetActorProcedure,
			opts...,
		),
		banActor: connect.NewClient[v1.BanActorRequest, v1.BanActorResponse](
			httpClient,
			baseURL+ModerationServiceBanActorProcedure,
			opts...,
		),
		unapproveActor: connect.NewClient[v1.UnapproveActorRequest, v1.UnapproveActorResponse](
			httpClient,
			baseURL+ModerationServiceUnapproveActorProcedure,
			opts...,
		),
		forceApproveActor: connect.NewClient[v1.ForceApproveActorRequest, v1.ForceApproveActorResponse](
			httpClient,
			baseURL+ModerationServiceForceApproveActorProcedure,
			opts...,
		),
		createActor: connect.NewClient[v1.CreateActorRequest, v1.CreateActorResponse](
			httpClient,
			baseURL+ModerationServiceCreateActorProcedure,
			opts...,
		),
		listAuditEvents: connect.NewClient[v1.ListAuditEventsRequest, v1.ListAuditEventsResponse](
			httpClient,
			baseURL+ModerationServiceListAuditEventsProcedure,
			opts...,
		),
		createCommentAuditEvent: connect.NewClient[v1.CreateCommentAuditEventRequest, v1.CreateCommentAuditEventResponse](
			httpClient,
			baseURL+ModerationServiceCreateCommentAuditEventProcedure,
			opts...,
		),
		listRoles: connect.NewClient[v1.ListRolesRequest, v1.ListRolesResponse](
			httpClient,
			baseURL+ModerationServiceListRolesProcedure,
			opts...,
		),
		assignRoles: connect.NewClient[v1.AssignRolesRequest, v1.AssignRolesResponse](
			httpClient,
			baseURL+ModerationServiceAssignRolesProcedure,
			opts...,
		),
	}
}

// moderationServiceClient implements ModerationServiceClient.
type moderationServiceClient struct {
	ping                    *connect.Client[v1.PingRequest, v1.PingResponse]
	processApprovalQueue    *connect.Client[v1.ProcessApprovalQueueRequest, v1.ProcessApprovalQueueResponse]
	holdBackPendingActor    *connect.Client[v1.HoldBackPendingActorRequest, v1.HoldBackPendingActorResponse]
	listActors              *connect.Client[v1.ListActorsRequest, v1.ListActorsResponse]
	getActor                *connect.Client[v1.GetActorRequest, v1.GetActorResponse]
	banActor                *connect.Client[v1.BanActorRequest, v1.BanActorResponse]
	unapproveActor          *connect.Client[v1.UnapproveActorRequest, v1.UnapproveActorResponse]
	forceApproveActor       *connect.Client[v1.ForceApproveActorRequest, v1.ForceApproveActorResponse]
	createActor             *connect.Client[v1.CreateActorRequest, v1.CreateActorResponse]
	listAuditEvents         *connect.Client[v1.ListAuditEventsRequest, v1.ListAuditEventsResponse]
	createCommentAuditEvent *connect.Client[v1.CreateCommentAuditEventRequest, v1.CreateCommentAuditEventResponse]
	listRoles               *connect.Client[v1.ListRolesRequest, v1.ListRolesResponse]
	assignRoles             *connect.Client[v1.AssignRolesRequest, v1.AssignRolesResponse]
}

// Ping calls bff.v1.ModerationService.Ping.
func (c *moderationServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// ProcessApprovalQueue calls bff.v1.ModerationService.ProcessApprovalQueue.
func (c *moderationServiceClient) ProcessApprovalQueue(ctx context.Context, req *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error) {
	return c.processApprovalQueue.CallUnary(ctx, req)
}

// HoldBackPendingActor calls bff.v1.ModerationService.HoldBackPendingActor.
func (c *moderationServiceClient) HoldBackPendingActor(ctx context.Context, req *connect.Request[v1.HoldBackPendingActorRequest]) (*connect.Response[v1.HoldBackPendingActorResponse], error) {
	return c.holdBackPendingActor.CallUnary(ctx, req)
}

// ListActors calls bff.v1.ModerationService.ListActors.
func (c *moderationServiceClient) ListActors(ctx context.Context, req *connect.Request[v1.ListActorsRequest]) (*connect.Response[v1.ListActorsResponse], error) {
	return c.listActors.CallUnary(ctx, req)
}

// GetActor calls bff.v1.ModerationService.GetActor.
func (c *moderationServiceClient) GetActor(ctx context.Context, req *connect.Request[v1.GetActorRequest]) (*connect.Response[v1.GetActorResponse], error) {
	return c.getActor.CallUnary(ctx, req)
}

// BanActor calls bff.v1.ModerationService.BanActor.
func (c *moderationServiceClient) BanActor(ctx context.Context, req *connect.Request[v1.BanActorRequest]) (*connect.Response[v1.BanActorResponse], error) {
	return c.banActor.CallUnary(ctx, req)
}

// UnapproveActor calls bff.v1.ModerationService.UnapproveActor.
func (c *moderationServiceClient) UnapproveActor(ctx context.Context, req *connect.Request[v1.UnapproveActorRequest]) (*connect.Response[v1.UnapproveActorResponse], error) {
	return c.unapproveActor.CallUnary(ctx, req)
}

// ForceApproveActor calls bff.v1.ModerationService.ForceApproveActor.
func (c *moderationServiceClient) ForceApproveActor(ctx context.Context, req *connect.Request[v1.ForceApproveActorRequest]) (*connect.Response[v1.ForceApproveActorResponse], error) {
	return c.forceApproveActor.CallUnary(ctx, req)
}

// CreateActor calls bff.v1.ModerationService.CreateActor.
func (c *moderationServiceClient) CreateActor(ctx context.Context, req *connect.Request[v1.CreateActorRequest]) (*connect.Response[v1.CreateActorResponse], error) {
	return c.createActor.CallUnary(ctx, req)
}

// ListAuditEvents calls bff.v1.ModerationService.ListAuditEvents.
func (c *moderationServiceClient) ListAuditEvents(ctx context.Context, req *connect.Request[v1.ListAuditEventsRequest]) (*connect.Response[v1.ListAuditEventsResponse], error) {
	return c.listAuditEvents.CallUnary(ctx, req)
}

// CreateCommentAuditEvent calls bff.v1.ModerationService.CreateCommentAuditEvent.
func (c *moderationServiceClient) CreateCommentAuditEvent(ctx context.Context, req *connect.Request[v1.CreateCommentAuditEventRequest]) (*connect.Response[v1.CreateCommentAuditEventResponse], error) {
	return c.createCommentAuditEvent.CallUnary(ctx, req)
}

// ListRoles calls bff.v1.ModerationService.ListRoles.
func (c *moderationServiceClient) ListRoles(ctx context.Context, req *connect.Request[v1.ListRolesRequest]) (*connect.Response[v1.ListRolesResponse], error) {
	return c.listRoles.CallUnary(ctx, req)
}

// AssignRoles calls bff.v1.ModerationService.AssignRoles.
func (c *moderationServiceClient) AssignRoles(ctx context.Context, req *connect.Request[v1.AssignRolesRequest]) (*connect.Response[v1.AssignRolesResponse], error) {
	return c.assignRoles.CallUnary(ctx, req)
}

// ModerationServiceHandler is an implementation of the bff.v1.ModerationService service.
type ModerationServiceHandler interface {
	// Ping is a test RPC that checks that the user is authenticated and then
	// returns an empty response. Ideal for health checking the moderation service.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// TODO: Refactor ProcessApprovalQueue to something more like "ProcessPendingActor"
	ProcessApprovalQueue(context.Context, *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error)
	// HoldBackPendingActor ignores a pending actor for review in some time, so we
	// don’t need to reject actors that e.g. have no avatar or bio yet.
	HoldBackPendingActor(context.Context, *connect.Request[v1.HoldBackPendingActorRequest]) (*connect.Response[v1.HoldBackPendingActorResponse], error)
	// ListActors fetches multiple actors from the database. It allows this to be
	// filtered by certain attributes.
	ListActors(context.Context, *connect.Request[v1.ListActorsRequest]) (*connect.Response[v1.ListActorsResponse], error)
	// GetActor fetches a single actor from the database.
	GetActor(context.Context, *connect.Request[v1.GetActorRequest]) (*connect.Response[v1.GetActorResponse], error)
	// BanActor changes an actors status to "banned".
	// Actor can be in any status before they are banned.
	BanActor(context.Context, *connect.Request[v1.BanActorRequest]) (*connect.Response[v1.BanActorResponse], error)
	// UnapproveActor changes an actor from "approved" status to "none" status.
	UnapproveActor(context.Context, *connect.Request[v1.UnapproveActorRequest]) (*connect.Response[v1.UnapproveActorResponse], error)
	// ForceApproveActor changes an actor to "approved" status from "none" or "pending".
	ForceApproveActor(context.Context, *connect.Request[v1.ForceApproveActorRequest]) (*connect.Response[v1.ForceApproveActorResponse], error)
	// CreateActor creates a database entry for an actor who does not currently exist.
	// By default, their status will be set to none.
	CreateActor(context.Context, *connect.Request[v1.CreateActorRequest]) (*connect.Response[v1.CreateActorResponse], error)
	ListAuditEvents(context.Context, *connect.Request[v1.ListAuditEventsRequest]) (*connect.Response[v1.ListAuditEventsResponse], error)
	CreateCommentAuditEvent(context.Context, *connect.Request[v1.CreateCommentAuditEventRequest]) (*connect.Response[v1.CreateCommentAuditEventResponse], error)
	ListRoles(context.Context, *connect.Request[v1.ListRolesRequest]) (*connect.Response[v1.ListRolesResponse], error)
	AssignRoles(context.Context, *connect.Request[v1.AssignRolesRequest]) (*connect.Response[v1.AssignRolesResponse], error)
}

// NewModerationServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewModerationServiceHandler(svc ModerationServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	moderationServicePingHandler := connect.NewUnaryHandler(
		ModerationServicePingProcedure,
		svc.Ping,
		opts...,
	)
	moderationServiceProcessApprovalQueueHandler := connect.NewUnaryHandler(
		ModerationServiceProcessApprovalQueueProcedure,
		svc.ProcessApprovalQueue,
		opts...,
	)
	moderationServiceHoldBackPendingActorHandler := connect.NewUnaryHandler(
		ModerationServiceHoldBackPendingActorProcedure,
		svc.HoldBackPendingActor,
		opts...,
	)
	moderationServiceListActorsHandler := connect.NewUnaryHandler(
		ModerationServiceListActorsProcedure,
		svc.ListActors,
		opts...,
	)
	moderationServiceGetActorHandler := connect.NewUnaryHandler(
		ModerationServiceGetActorProcedure,
		svc.GetActor,
		opts...,
	)
	moderationServiceBanActorHandler := connect.NewUnaryHandler(
		ModerationServiceBanActorProcedure,
		svc.BanActor,
		opts...,
	)
	moderationServiceUnapproveActorHandler := connect.NewUnaryHandler(
		ModerationServiceUnapproveActorProcedure,
		svc.UnapproveActor,
		opts...,
	)
	moderationServiceForceApproveActorHandler := connect.NewUnaryHandler(
		ModerationServiceForceApproveActorProcedure,
		svc.ForceApproveActor,
		opts...,
	)
	moderationServiceCreateActorHandler := connect.NewUnaryHandler(
		ModerationServiceCreateActorProcedure,
		svc.CreateActor,
		opts...,
	)
	moderationServiceListAuditEventsHandler := connect.NewUnaryHandler(
		ModerationServiceListAuditEventsProcedure,
		svc.ListAuditEvents,
		opts...,
	)
	moderationServiceCreateCommentAuditEventHandler := connect.NewUnaryHandler(
		ModerationServiceCreateCommentAuditEventProcedure,
		svc.CreateCommentAuditEvent,
		opts...,
	)
	moderationServiceListRolesHandler := connect.NewUnaryHandler(
		ModerationServiceListRolesProcedure,
		svc.ListRoles,
		opts...,
	)
	moderationServiceAssignRolesHandler := connect.NewUnaryHandler(
		ModerationServiceAssignRolesProcedure,
		svc.AssignRoles,
		opts...,
	)
	return "/bff.v1.ModerationService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ModerationServicePingProcedure:
			moderationServicePingHandler.ServeHTTP(w, r)
		case ModerationServiceProcessApprovalQueueProcedure:
			moderationServiceProcessApprovalQueueHandler.ServeHTTP(w, r)
		case ModerationServiceHoldBackPendingActorProcedure:
			moderationServiceHoldBackPendingActorHandler.ServeHTTP(w, r)
		case ModerationServiceListActorsProcedure:
			moderationServiceListActorsHandler.ServeHTTP(w, r)
		case ModerationServiceGetActorProcedure:
			moderationServiceGetActorHandler.ServeHTTP(w, r)
		case ModerationServiceBanActorProcedure:
			moderationServiceBanActorHandler.ServeHTTP(w, r)
		case ModerationServiceUnapproveActorProcedure:
			moderationServiceUnapproveActorHandler.ServeHTTP(w, r)
		case ModerationServiceForceApproveActorProcedure:
			moderationServiceForceApproveActorHandler.ServeHTTP(w, r)
		case ModerationServiceCreateActorProcedure:
			moderationServiceCreateActorHandler.ServeHTTP(w, r)
		case ModerationServiceListAuditEventsProcedure:
			moderationServiceListAuditEventsHandler.ServeHTTP(w, r)
		case ModerationServiceCreateCommentAuditEventProcedure:
			moderationServiceCreateCommentAuditEventHandler.ServeHTTP(w, r)
		case ModerationServiceListRolesProcedure:
			moderationServiceListRolesHandler.ServeHTTP(w, r)
		case ModerationServiceAssignRolesProcedure:
			moderationServiceAssignRolesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedModerationServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedModerationServiceHandler struct{}

func (UnimplementedModerationServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.Ping is not implemented"))
}

func (UnimplementedModerationServiceHandler) ProcessApprovalQueue(context.Context, *connect.Request[v1.ProcessApprovalQueueRequest]) (*connect.Response[v1.ProcessApprovalQueueResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.ProcessApprovalQueue is not implemented"))
}

func (UnimplementedModerationServiceHandler) HoldBackPendingActor(context.Context, *connect.Request[v1.HoldBackPendingActorRequest]) (*connect.Response[v1.HoldBackPendingActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.HoldBackPendingActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) ListActors(context.Context, *connect.Request[v1.ListActorsRequest]) (*connect.Response[v1.ListActorsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.ListActors is not implemented"))
}

func (UnimplementedModerationServiceHandler) GetActor(context.Context, *connect.Request[v1.GetActorRequest]) (*connect.Response[v1.GetActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.GetActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) BanActor(context.Context, *connect.Request[v1.BanActorRequest]) (*connect.Response[v1.BanActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.BanActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) UnapproveActor(context.Context, *connect.Request[v1.UnapproveActorRequest]) (*connect.Response[v1.UnapproveActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.UnapproveActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) ForceApproveActor(context.Context, *connect.Request[v1.ForceApproveActorRequest]) (*connect.Response[v1.ForceApproveActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.ForceApproveActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) CreateActor(context.Context, *connect.Request[v1.CreateActorRequest]) (*connect.Response[v1.CreateActorResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.CreateActor is not implemented"))
}

func (UnimplementedModerationServiceHandler) ListAuditEvents(context.Context, *connect.Request[v1.ListAuditEventsRequest]) (*connect.Response[v1.ListAuditEventsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.ListAuditEvents is not implemented"))
}

func (UnimplementedModerationServiceHandler) CreateCommentAuditEvent(context.Context, *connect.Request[v1.CreateCommentAuditEventRequest]) (*connect.Response[v1.CreateCommentAuditEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.CreateCommentAuditEvent is not implemented"))
}

func (UnimplementedModerationServiceHandler) ListRoles(context.Context, *connect.Request[v1.ListRolesRequest]) (*connect.Response[v1.ListRolesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.ListRoles is not implemented"))
}

func (UnimplementedModerationServiceHandler) AssignRoles(context.Context, *connect.Request[v1.AssignRolesRequest]) (*connect.Response[v1.AssignRolesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.ModerationService.AssignRoles is not implemented"))
}
