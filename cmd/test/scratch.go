package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
)

func main() {
	run("https://bsky.network")
	run("https://bsky.social")
}

func run(host string) {
	ctx := context.Background()
	x := &xrpc.Client{
		Host: host,
	}
	myDID := "did:plc:445avk3am7zpwlrj7aop746e"
	//out, err := atproto.SyncGetLatestCommit(ctx, x, myDID)
	//if err != nil {
	//	panic(err)
	//}

	rootCommitFromGetRecord := ""
	revFromGetRecord := ""
	{
		data, err := atproto.SyncGetRecord(ctx, x, "app.bsky.actor.profile", "", myDID, "self")
		if err != nil {
			panic(err)
		}
		bs := blockstore.NewBlockstore(datastore.NewMapDatastore())
		root, err := repo.IngestRepo(ctx, bs, bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		rr, err := repo.OpenRepo(ctx, bs, root, false)
		if err != nil {
			panic(err)
		}
		revFromGetRecord = rr.SignedCommit().Rev
		rootCommitFromGetRecord = root.String()
	}

	rootCommitFromGetRepo := ""
	revFromGetRepo := ""
	{
		data, err := atproto.SyncGetRepo(ctx, x, myDID, "")
		if err != nil {
			panic(err)
		}
		bs := blockstore.NewBlockstore(datastore.NewMapDatastore())
		root, err := repo.IngestRepo(ctx, bs, bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		rr, err := repo.OpenRepo(ctx, bs, root, false)
		if err != nil {
			panic(err)
		}
		revFromGetRepo = rr.SignedCommit().Rev
		rootCommitFromGetRepo = root.String()
	}

	fmt.Println("-- " + host + " --")
	//fmt.Println("SyncGetLatestCommit root commit: " + out.Cid)
	fmt.Println("SyncGetRecord root commit:       " + rootCommitFromGetRecord)
	fmt.Println("SyncGetRepo root commit:         " + rootCommitFromGetRepo)
	//fmt.Println("SyncGetLatestCommit rev: " + out.Rev)
	fmt.Println("SyncGetRecord rev:       " + revFromGetRecord)
	fmt.Println("SyncGetRepo rev:         " + revFromGetRepo)
}
