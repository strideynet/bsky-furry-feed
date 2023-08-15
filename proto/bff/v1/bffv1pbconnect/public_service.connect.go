// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: bff/v1/public_service.proto

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
	// PublicServiceName is the fully-qualified name of the PublicService service.
	PublicServiceName = "bff.v1.PublicService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// PublicServiceListFeedsProcedure is the fully-qualified name of the PublicService's ListFeeds RPC.
	PublicServiceListFeedsProcedure = "/bff.v1.PublicService/ListFeeds"
)

// PublicServiceClient is a client for the bff.v1.PublicService service.
type PublicServiceClient interface {
	// ListFeeds returns a list of all feeds hosted by this server.
	ListFeeds(context.Context, *connect.Request[v1.ListFeedsRequest]) (*connect.Response[v1.ListFeedsResponse], error)
}

// NewPublicServiceClient constructs a client for the bff.v1.PublicService service. By default, it
// uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewPublicServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) PublicServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &publicServiceClient{
		listFeeds: connect.NewClient[v1.ListFeedsRequest, v1.ListFeedsResponse](
			httpClient,
			baseURL+PublicServiceListFeedsProcedure,
			opts...,
		),
	}
}

// publicServiceClient implements PublicServiceClient.
type publicServiceClient struct {
	listFeeds *connect.Client[v1.ListFeedsRequest, v1.ListFeedsResponse]
}

// ListFeeds calls bff.v1.PublicService.ListFeeds.
func (c *publicServiceClient) ListFeeds(ctx context.Context, req *connect.Request[v1.ListFeedsRequest]) (*connect.Response[v1.ListFeedsResponse], error) {
	return c.listFeeds.CallUnary(ctx, req)
}

// PublicServiceHandler is an implementation of the bff.v1.PublicService service.
type PublicServiceHandler interface {
	// ListFeeds returns a list of all feeds hosted by this server.
	ListFeeds(context.Context, *connect.Request[v1.ListFeedsRequest]) (*connect.Response[v1.ListFeedsResponse], error)
}

// NewPublicServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewPublicServiceHandler(svc PublicServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	publicServiceListFeedsHandler := connect.NewUnaryHandler(
		PublicServiceListFeedsProcedure,
		svc.ListFeeds,
		opts...,
	)
	return "/bff.v1.PublicService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PublicServiceListFeedsProcedure:
			publicServiceListFeedsHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedPublicServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedPublicServiceHandler struct{}

func (UnimplementedPublicServiceHandler) ListFeeds(context.Context, *connect.Request[v1.ListFeedsRequest]) (*connect.Response[v1.ListFeedsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bff.v1.PublicService.ListFeeds is not implemented"))
}
