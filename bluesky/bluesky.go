package bluesky

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

type UnauthClient struct {
	xrpc *xrpc.Client
}

type Client struct {
	*UnauthClient
	xrpc     *xrpc.Client
	authInfo *xrpc.AuthInfo
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
	return atproto.IdentityResolveHandle(
		ctx,
		c.xrpc,
		handle,
	)
}

func (c *UnauthClient) CreateSession(ctx context.Context, identifier string, password string) (*atproto.ServerCreateSession_Output, error) {
	return atproto.ServerCreateSession(
		ctx,
		c.xrpc,
		&atproto.ServerCreateSession_Input{
			Identifier: identifier,
			Password:   password,
		},
	)
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
		authInfo: auth,
	}
}

func (c *Client) GetFollowers(
	ctx context.Context, actor string, cursor string, limit int64,
) (*bsky.GraphGetFollowers_Output, error) {
	return bsky.GraphGetFollowers(
		ctx,
		c.xrpc,
		actor,
		cursor,
		limit,
	)
}

func (c *Client) GetPostThread(
	ctx context.Context, uri string,
) (*bsky.FeedGetPostThread_Output, error) {
	return bsky.FeedGetPostThread(ctx, c.xrpc, 1, uri)
}

// GetProfile fetches an actor's profile. actor can be a DID or a handle.
func (c *Client) GetProfile(
	ctx context.Context, actor string,
) (*bsky.ActorDefs_ProfileViewDetailed, error) {
	return bsky.ActorGetProfile(ctx, c.xrpc, actor)
}

// Follow creates an app.bsky.graph.follow for the user the client is
// authenticated as.
func (c *Client) Follow(
	ctx context.Context, subjectDID string,
) error {
	// {
	// 	"collection":"app.bsky.graph.follow",
	//	"repo":"did:plc:jdkvwye2lf4mingzk7qdebzc",
	//	"record":{
	//		"subject":"did:plc:nb5a2kg3gnrxe5wrw47grzac",
	//		"createdAt":"2023-05-21T12:47:14.733Z",
	//		"$type":"app.bsky.graph.follow"
	//	}
	// }
	createRecord := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       c.authInfo.Did,
		Record: &util.LexiconTypeDecoder{
			Val: &bsky.GraphFollow{
				CreatedAt: FormatTime(time.Now()),
				Subject:   subjectDID,
			},
		},
	}
	_, err := atproto.RepoCreateRecord(ctx, c.xrpc, createRecord)
	if err != nil {
		return err
	}
	return nil
}

// GetSession calls com.atproto.server.getSession
func (c *Client) GetSession(ctx context.Context) (*atproto.ServerGetSession_Output, error) {
	return atproto.ServerGetSession(ctx, c.xrpc)
}

var ErrMalformedRecordUri = fmt.Errorf("malformed record uri")

// Parse the namespace ID from a full record URI, such
// as app.bsky.feed.post or app.bsky.graph.follow.
//
// Errors with a `ErrMalformedRecordUri` if the URI is
// not a _parseable_ record URI.
func ParseNamespaceID(recordUri string) (string, error) {
	parts := strings.Split(recordUri, "/")

	if len(parts) <= 3 {
		return "", ErrMalformedRecordUri
	}

	return parts[3], nil
}
