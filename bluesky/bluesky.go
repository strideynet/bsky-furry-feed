package bluesky

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

type Client struct {
	xrpc *xrpc.Client
}

type Credentials struct {
	Identifier string
	Password   string
}

func CredentialsFromEnv() (*Credentials, error) {
	identifier := os.Getenv("BLUESKY_USERNAME")
	if identifier == "" {
		return nil, fmt.Errorf("BLUESKY_USERNAME environment variable not set")
	}
	password := os.Getenv("BLUESKY_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("BLUESKY_PASSWORD environment variable not set")
	}

	return &Credentials{Identifier: identifier, Password: password}, nil
}

func baseXRPC() *xrpc.Client {
	// TODO: Introduce a ClientConfig we can control these with
	ua := "github.com/strideynet/bluesky-furry-feed"
	return &xrpc.Client{
		Host:      "https://bsky.social",
		UserAgent: &ua,
	}
}

func ClientFromCredentials(ctx context.Context, credentials *Credentials) (*Client, error) {
	xrpcClient := baseXRPC()
	out, err := atproto.ServerCreateSession(
		ctx,
		xrpcClient,
		&atproto.ServerCreateSession_Input{
			Identifier: credentials.Identifier,
			Password:   credentials.Password,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	xrpcClient.Auth = &xrpc.AuthInfo{
		AccessJwt:  out.AccessJwt,
		RefreshJwt: out.RefreshJwt,
		Did:        out.Did,
		Handle:     out.Handle,
	}

	return &Client{xrpc: xrpcClient}, nil
}

// ClientFromToken takes a JWT access token, and makes a client. It then calls
// GetSession to verify the token.
//
// On success, an authenticated client is returned along with the JWTs DID
func ClientFromToken(ctx context.Context, token string) (*Client, string, error) {
	xrpcClient := baseXRPC()
	xrpcClient.Auth = &xrpc.AuthInfo{
		AccessJwt: token,
	}

	res, err := atproto.ServerGetSession(ctx, xrpcClient)
	if err != nil {
		return nil, "", fmt.Errorf("fetching session: %w", err)
	}
	xrpcClient.Auth.Did = res.Did
	xrpcClient.Auth.Handle = res.Handle

	return &Client{xrpc: xrpcClient}, res.Did, nil
}

func (c *Client) ResolveHandle(ctx context.Context, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	return atproto.IdentityResolveHandle(
		ctx,
		c.xrpc,
		handle,
	)
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
		Repo:       c.xrpc.Auth.Did,
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
