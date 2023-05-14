package bluesky

import (
	"context"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
)

type Client struct {
	xrpc *xrpc.Client
}

func NewClient() *Client {
	return &Client{
		xrpc: &xrpc.Client{
			Host: "https://bsky.social",
		},
	}
}

func (c *Client) ResolveHandle(ctx context.Context, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	did, err := atproto.IdentityResolveHandle(
		ctx,
		c.xrpc,
		handle,
	)
	if err != nil {
		return nil, err
	}
	return did, nil
}
