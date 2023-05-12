package main

import (
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "bffctl",
		Usage: "The swiss army knife of any BFF operator",
		Commands: []*cli.Command{
			dbCmd(),
			findDID(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func findDID() *cli.Command {
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
			pdsClient := &xrpc.Client{
				Host: "https://bsky.social",
			}
			did, err := atproto.IdentityResolveHandle(
				cctx.Context,
				pdsClient,
				handle,
			)
			if err != nil {
				return err
			}
			fmt.Printf("Found DID: %s\n", did.Did)
			return nil
		},
	}
}
