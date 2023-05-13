package main

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"
)

func main() {
	log, _ := zap.NewDevelopment()
	app := &cli.App{
		Name:  "bffctl",
		Usage: "The swiss army knife of any BFF operator",
		Commands: []*cli.Command{
			dbCmd(log),
			findDIDCmd(log),
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func findDID(ctx context.Context, handle string) (string, error) {
	pdsClient := &xrpc.Client{
		Host: "https://bsky.social",
	}
	did, err := atproto.IdentityResolveHandle(
		ctx,
		pdsClient,
		handle,
	)
	if err != nil {
		return "", err
	}
	return did.Did, nil
}

func findDIDCmd(log *zap.Logger) *cli.Command {
	handle := ""
	return &cli.Command{
		Name: "find-did",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "handle",
				Usage:       "Find DID for handle",
				Destination: &handle,
				Required:    true,
			},
		},
		Action: func(cctx *cli.Context) error {
			did, err := findDID(cctx.Context, handle)
			if err != nil {
				return err
			}
			log.Info("found did", zap.String("did", did))
			return nil
		},
	}
}
