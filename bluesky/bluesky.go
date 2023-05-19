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

func (c *Client) GetFollowers(
	ctx context.Context, actor string, cursor string, limit int64,
) (*bsky.GraphGetFollowers_Output, error) {
	out, err := bsky.GraphGetFollowers(
		ctx,
		c.xrpc,
		actor,
		cursor,
		limit,
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetPostThread(
	ctx context.Context, uri string,
) (*bsky.FeedGetPostThread_Output, error) {
	out, err := bsky.FeedGetPostThread(ctx, c.xrpc, 1, uri)
	if err != nil {
		return nil, err
	}
	return out, err
}

// GetProfile fetches an actor's profile. actor can be a DID or a handle.
func (c *Client) GetProfile(
	ctx context.Context, actor string,
) (*bsky.ActorDefs_ProfileViewDetailed, error) {
	out, err := bsky.ActorGetProfile(ctx, c.xrpc, actor)
	if err != nil {
		return nil, err
	}
	return out, err
}
