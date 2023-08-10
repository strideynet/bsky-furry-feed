package bluesky

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	indigoUtils "github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/ipfs/go-cid"
	typegen "github.com/whyrusleeping/cbor-gen"
)

const DefaultPDSHost = "https://bsky.social"

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

func baseXRPC(pdsHost string) *xrpc.Client {
	// TODO: Introduce a ClientConfig we can control these with
	ua := "github.com/strideynet/bluesky-furry-feed"
	return &xrpc.Client{
		Host:      pdsHost,
		UserAgent: &ua,
	}
}

func ClientFromCredentials(ctx context.Context, pdsHost string, credentials *Credentials) (*Client, error) {
	xrpcClient := baseXRPC(pdsHost)
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
func ClientFromToken(ctx context.Context, pdsHost string, token string) (*Client, string, error) {
	xrpcClient := baseXRPC(pdsHost)
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

// GetProfile fetches an actor's profile. actor can be a DID or a handle.
func (c *Client) GetProfile(
	ctx context.Context, actor string,
) (*bsky.ActorDefs_ProfileViewDetailed, error) {
	return bsky.ActorGetProfile(ctx, c.xrpc, actor)
}

func (c *Client) GetHead(
	ctx context.Context, actorDID string,
) (cid.Cid, error) {
	resp, err := atproto.SyncGetHead(ctx, c.xrpc, actorDID)
	if err != nil {
		return cid.Cid{}, err
	}
	out, err := cid.Parse(resp.Root)
	if err != nil {
		return cid.Cid{}, err
	}
	return out, nil
}

func (c *Client) GetRecord(
	ctx context.Context, collection string, commitCID cid.Cid, actorDID string, rkey string,
) (typegen.CBORMarshaler, error) {
	// We can't use RepoGetRecord here, because RepoGetRecord gets the record by the record's CID and not the commit's CID.
	blocks, err := atproto.SyncGetRecord(ctx, c.xrpc, collection, commitCID.String(), actorDID, rkey)
	if err != nil {
		return nil, err
	}

	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(blocks))
	if err != nil {
		return nil, err
	}

	_, record, err := rr.GetRecord(ctx, collection+"/"+rkey)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// Follow creates an app.bsky.graph.follow for the user the client is
// authenticated as.
func (c *Client) Follow(
	ctx context.Context, subjectDID string,
) error {
	profile, err := c.GetProfile(ctx, subjectDID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	if profile.Viewer.Following != nil {
		// Already following - no need to follow.
		return nil
	}

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
	_, err = atproto.RepoCreateRecord(ctx, c.xrpc, createRecord)
	if err != nil {
		return fmt.Errorf("creating follow record: %w", err)
	}
	return nil
}

// Unfollow removes any app.bsky.graph.follow for the subject from the account
// the client is authenticated as.
func (c *Client) Unfollow(
	ctx context.Context, subjectDID string,
) error {
	profile, err := c.GetProfile(ctx, subjectDID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	if profile.Viewer.Following == nil {
		// Nothing to unfollow
		return nil
	}

	uri, err := indigoUtils.ParseAtUri(*profile.Viewer.Following)
	if err != nil {
		return fmt.Errorf("parsing following uri: %w", err)
	}

	err = atproto.RepoDeleteRecord(ctx, c.xrpc, &atproto.RepoDeleteRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       c.xrpc.Auth.Did,
		Rkey:       uri.Rkey,
	})
	if err != nil {
		return fmt.Errorf("deleting follow record: %w", err)
	}
	return nil
}
