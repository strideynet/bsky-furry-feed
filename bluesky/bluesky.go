package bluesky

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	indigoUtils "github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ipfs/go-cid"
	typegen "github.com/whyrusleeping/cbor-gen"
)

const DefaultPDSHost = "https://bsky.social"

type tokenInfo struct {
	authInfo  *xrpc.AuthInfo
	expiresAt time.Time
}

func tokenInfoFromAuthInfo(authInfo *xrpc.AuthInfo) (tokenInfo, error) {
	var claims jwt.RegisteredClaims
	if _, _, err := jwt.NewParser().ParseUnverified(authInfo.AccessJwt, &claims); err != nil {
		return tokenInfo{}, fmt.Errorf("failed to parse jwt: %w", err)
	}

	return tokenInfo{
		authInfo:  authInfo,
		expiresAt: claims.ExpiresAt.Time,
	}, nil
}

type Client struct {
	pdsHost string

	tokenInfo   tokenInfo
	tokenInfoMu sync.Mutex
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

func ClientFromCredentials(ctx context.Context, pdsHost string, credentials *Credentials) (*Client, error) {
	c := &Client{
		pdsHost: pdsHost,
	}

	sess, err := atproto.ServerCreateSession(
		ctx,
		c.baseXRPCClient(),
		&atproto.ServerCreateSession_Input{
			Identifier: credentials.Identifier,
			Password:   credentials.Password,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	ti, err := tokenInfoFromAuthInfo(&xrpc.AuthInfo{
		AccessJwt:  sess.AccessJwt,
		RefreshJwt: sess.RefreshJwt,
		Did:        sess.Did,
		Handle:     sess.Handle,
	})
	if err != nil {
		return nil, err
	}

	// c.tokenInfoMu does not need to be locked here on first initialization.
	c.tokenInfo = ti

	return c, nil
}

const UserAgent = "github.com/strideynet/bluesky-furry-feed"

func (c *Client) baseXRPCClient() *xrpc.Client {
	// TODO: Introduce a ClientConfig we can control these with
	ua := UserAgent
	return &xrpc.Client{
		Host:      c.pdsHost,
		UserAgent: &ua,
	}
}

func (c *Client) xrpcClient(ctx context.Context) (*xrpc.Client, error) {
	c.tokenInfoMu.Lock()
	defer c.tokenInfoMu.Unlock()

	if time.Now().After(c.tokenInfo.expiresAt.Add(-10 * time.Minute)) {
		if err := c.refreshToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	xc := c.baseXRPCClient()
	xc.Auth = &xrpc.AuthInfo{
		AccessJwt:  c.tokenInfo.authInfo.AccessJwt,
		RefreshJwt: c.tokenInfo.authInfo.RefreshJwt,
		Handle:     c.tokenInfo.authInfo.Handle,
		Did:        c.tokenInfo.authInfo.Did,
	}
	return xc, nil
}

func (c *Client) refreshToken(ctx context.Context) error {
	xc := c.baseXRPCClient()
	xc.Auth = &xrpc.AuthInfo{
		AccessJwt: c.tokenInfo.authInfo.RefreshJwt,
	}

	sess, err := atproto.ServerRefreshSession(ctx, xc)
	if err != nil {
		return fmt.Errorf("refresh session: %w", err)
	}

	ti, err := tokenInfoFromAuthInfo(&xrpc.AuthInfo{
		AccessJwt:  sess.AccessJwt,
		RefreshJwt: sess.RefreshJwt,
		Did:        sess.Did,
		Handle:     sess.Handle,
	})
	if err != nil {
		return err
	}

	c.tokenInfo = ti
	return nil
}

func (c *Client) ResolveHandle(ctx context.Context, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}
	return atproto.IdentityResolveHandle(ctx, xc, handle)
}

func (c *Client) GetFollowers(
	ctx context.Context, actor string, cursor string, limit int64,
) (*bsky.GraphGetFollowers_Output, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}
	return bsky.GraphGetFollowers(ctx, xc, actor, cursor, limit)
}

// GetProfile fetches an actor's profile. actor can be a DID or a handle.
func (c *Client) GetProfile(
	ctx context.Context, actor string,
) (*bsky.ActorDefs_ProfileViewDetailed, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}
	return bsky.ActorGetProfile(ctx, xc, actor)
}

func (c *Client) GetHead(
	ctx context.Context, actorDID string,
) (cid.Cid, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("get xrpc client: %w", err)
	}
	resp, err := atproto.SyncGetHead(ctx, xc, actorDID)
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
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}

	// We can't use RepoGetRecord here, because RepoGetRecord gets the record by the record's CID and not the commit's CID.
	blocks, err := atproto.SyncGetRecord(ctx, xc, collection, commitCID.String(), actorDID, rkey)
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
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return fmt.Errorf("get xrpc client: %w", err)
	}

	profile, err := bsky.ActorGetProfile(ctx, xc, subjectDID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	if profile.Viewer.Following != nil {
		// Already following - no need to follow.
		return nil
	}

	createRecord := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       xc.Auth.Did,
		Record: &util.LexiconTypeDecoder{
			Val: &bsky.GraphFollow{
				CreatedAt: FormatTime(time.Now()),
				Subject:   subjectDID,
			},
		},
	}
	_, err = atproto.RepoCreateRecord(ctx, xc, createRecord)
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

	err = c.DeleteRecord(ctx, uri)
	if err != nil {
		return fmt.Errorf("deleting follow record: %w", err)
	}
	return nil
}

// DeleteRecord deletes a record from a repository
func (c *Client) DeleteRecord(
	ctx context.Context, uri *indigoUtils.ParsedUri,
) error {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return fmt.Errorf("get xrpc client: %w", err)
	}

	if err := atproto.RepoDeleteRecord(ctx, xc, &atproto.RepoDeleteRecord_Input{
		Collection: uri.Collection,
		Repo:       uri.Did,
		Rkey:       uri.Rkey,
	}); err != nil {
		return fmt.Errorf("deleting record: %w", err)
	}
	return nil
}

// PurgeFeeds deletes all feeds associated with the authenticated user
func (c *Client) PurgeFeeds(
	ctx context.Context,
) error {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return fmt.Errorf("get xrpc client: %w", err)
	}

	// TODO: Pagination
	out, err := bsky.FeedGetActorFeeds(ctx, xc, xc.Auth.Did, "", 100)
	if err != nil {
		return fmt.Errorf("getting feeds: %w", err)
	}

	for _, f := range out.Feeds {
		uri, err := indigoUtils.ParseAtUri(f.Uri)
		if err != nil {
			return fmt.Errorf("parsing feed uri: %w", err)
		}
		err = c.DeleteRecord(ctx, uri)
		if err != nil {
			return fmt.Errorf("deleting record: %w", err)
		}
	}

	return nil
}

// This exists because the go code gen is incorrect for swapRecord and misses
// an omitEmpty on SwapRecord.
// putting feed record: putting record: XRPC ERROR 400: InvalidSwap: Record was at bafyreigkeuzjkpot7yzpseezz4hat2jmlobypfhtaaisxbdlwafwxp4ywa
type RepoPutRecord_Input struct {
	// collection: The NSID of the record collection.
	Collection string `json:"collection" cborgen:"collection"`
	// record: The record to write.
	Record *util.LexiconTypeDecoder `json:"record" cborgen:"record"`
	// repo: The handle or DID of the repo.
	Repo string `json:"repo" cborgen:"repo"`
	// rkey: The key of the record.
	Rkey string `json:"rkey" cborgen:"rkey"`
	// swapCommit: Compare and swap with the previous commit by cid.
	SwapCommit *string `json:"swapCommit,omitempty" cborgen:"swapCommit,omitempty"`
	// swapRecord: Compare and swap with the previous record by cid.
	SwapRecord *string `json:"swapRecord,omitempty" cborgen:"swapRecord,omitempty"`
	// validate: Validate the record?
	Validate *bool `json:"validate,omitempty" cborgen:"validate,omitempty"`
}

// PutRecord creates or updates a record in the actor's repository.
func (c *Client) PutRecord(
	ctx context.Context, collection, rkey string, record repo.CborMarshaler,
) error {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return fmt.Errorf("get xrpc client: %w", err)
	}

	var out atproto.RepoPutRecord_Output
	if err := xc.Do(ctx, xrpc.Procedure, "application/json", "com.atproto.repo.putRecord", nil, &RepoPutRecord_Input{
		Collection: collection,
		Repo:       xc.Auth.Did,
		Rkey:       rkey,
		Record: &util.LexiconTypeDecoder{
			Val: record,
		},
	}, &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) UploadBlob(
	ctx context.Context, blob io.Reader,
) (*util.LexBlob, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}

	// set encoding: 'image/png'
	out, err := atproto.RepoUploadBlob(ctx, xc, blob)
	if err != nil {
		return nil, fmt.Errorf("uploading blob: %w", err)
	}
	return out.Blob, nil
}
