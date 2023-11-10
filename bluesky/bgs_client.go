package bluesky

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/xrpc"
	typegen "github.com/whyrusleeping/cbor-gen"
)

const DefaultBGSHost = "https://bsky.network"

type BGSClient struct {
	BGSHost string
}

func (c *BGSClient) xrpcClient() *xrpc.Client {
	ua := UserAgent
	host := c.BGSHost
	if host == "" {
		host = DefaultBGSHost
	}
	return &xrpc.Client{
		Host:      host,
		UserAgent: &ua,
	}
}

// SyncGetRecord invokes the `SyncGetRecord` RPC against the BGS, and then
// parses the returned CAR to retrieve the record and the current repo rev.
func (c *BGSClient) SyncGetRecord(
	ctx context.Context, collection string, actorDID string, rkey string,
) (record typegen.CBORMarshaler, repoRev string, err error) {
	xc := c.xrpcClient()

	blocks, err := atproto.SyncGetRecord(ctx, xc, collection, "", actorDID, rkey)
	if err != nil {
		return nil, "", fmt.Errorf("calling SyncGetRecord: %w", err)
	}

	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(blocks))
	if err != nil {
		return nil, "", fmt.Errorf("reading repo from car: %w", err)
	}

	_, record, err = rr.GetRecord(ctx, collection+"/"+rkey)
	if err != nil {
		return nil, "", fmt.Errorf("getting record: %w", err)
	}

	return record, rr.SignedCommit().Rev, nil
}
