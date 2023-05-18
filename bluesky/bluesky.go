package bluesky

import (
	"context"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

type UnauthClient struct {
	xrpc *xrpc.Client
}

type Client struct {
	*UnauthClient
	xrpc *xrpc.Client
}

var userAgent = "github.com/strideynet/bluesky-furry-feed"
var host = "https://bsky.social"

func NewUnauthClient() *UnauthClient {
	return &UnauthClient{
		xrpc: &xrpc.Client{
			Host:      host,
			UserAgent: &userAgent,
		},
	}
}

func (c *UnauthClient) ResolveHandle(ctx context.Context, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	out, err := atproto.IdentityResolveHandle(
		ctx,
		c.xrpc,
		handle,
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *UnauthClient) CreateSession(ctx context.Context, identifier string, password string) (*atproto.ServerCreateSession_Output, error) {
	out, err := atproto.ServerCreateSession(
		ctx,
		c.xrpc,
		&atproto.ServerCreateSession_Input{
			Identifier: identifier,
			Password:   password,
		},
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func AuthInfoFromCreateSession(in *atproto.ServerCreateSession_Output) *xrpc.AuthInfo {
	return &xrpc.AuthInfo{
		AccessJwt:  in.AccessJwt,
		RefreshJwt: in.RefreshJwt,
		Did:        in.Did,
		Handle:     in.Handle,
	}
}

// TODO: Manage refreshing identity for user and provide some way to persist
// this refreshed identity.
func NewClient(auth *xrpc.AuthInfo) *Client {
	return &Client{
		UnauthClient: NewUnauthClient(),
		xrpc: &xrpc.Client{
			Host:      host,
			Auth:      auth,
			UserAgent: &userAgent,
		},
	}
}

func (c *Client) GetFollowers(ctx context.Context, actor string) (*atproto.IdentityResolveHandle_Output, error) {
	bsky.GraphGetFollowers(ctx, c.xrpc, actor, "", 100)
	out, err := atproto.IdentityResolveHandle(
		ctx,
		c.xrpc,
		handle,
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}
