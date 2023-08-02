//go:build tools

package bff

import (
	_ "connectrpc.com/connect/cmd/protoc-gen-connect-go"
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
)
