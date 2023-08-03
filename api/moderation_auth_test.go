package api

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/proto"
	"testing"
)

type memoryActorGetter map[string]*v1.Actor

func (mag memoryActorGetter) GetActorByDID(_ context.Context, did string) (*v1.Actor, error) {
	v, ok := mag[did]
	if !ok {
		// TODO: We really ought to have a store.NotFound
		return nil, pgx.ErrNoRows
	}
	return proto.Clone(v).(*v1.Actor), nil
}

func TestAuthEngine(t *testing.T) {
	tests := []struct {
		name string

		actorGetter   actorGetter
		procedureName string

		want    *authContext
		wantErr string
	}{
		{
			name: "success",
			actorGetter: memoryActorGetter{
				"exists": &v1.Actor{
					Did:   "exists",
					Roles: []string{"admin"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ae := &AuthEngine{
				ActorGetter: tt.actorGetter,
				// TODO: Inject TokenValidator fake
				Log: zaptest.NewLogger(t),
			}

			req := connect.NewRequest(&v1.PingRequest{})
			// TODO: figure out how to inject a procedureName
			got, err := ae.auth(ctx, req)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
