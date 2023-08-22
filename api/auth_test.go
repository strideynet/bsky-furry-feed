package api

import (
	"connectrpc.com/connect"
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"reflect"
	"testing"
	"unsafe"
)

type memoryActorGetter map[string]*v1.Actor

func (mag memoryActorGetter) GetActorByDID(_ context.Context, did string) (*v1.Actor, error) {
	v, ok := mag[did]
	if !ok {
		return nil, store.ErrNotFound
	}
	return proto.Clone(v).(*v1.Actor), nil
}

func setSpec(req connect.AnyRequest, spec connect.Spec) {
	specVal := reflect.ValueOf(req).
		Elem().FieldByName("spec")

	reflect.NewAt(
		specVal.Type(),
		unsafe.Pointer(specVal.UnsafeAddr())).Elem().Set(reflect.ValueOf(spec))
}

func TestAuthEngine(t *testing.T) {
	tests := []struct {
		name string

		headerKey     string
		headerValue   string
		actor         *v1.Actor
		procedureName string

		want    *authContext
		wantErr string
	}{
		{
			name:          "success",
			headerKey:     "Authorization",
			headerValue:   "Bearer exists",
			procedureName: "/bff.v1.ModerationService/CreateActor",
			actor: &v1.Actor{
				Did:   "exists",
				Roles: []string{"admin"},
			},
			want: &authContext{
				DID: "exists",
				Actor: &v1.Actor{
					Did:   "exists",
					Roles: []string{"admin"},
				},
			},
		},
		{
			name:          "success: non-existent user",
			headerKey:     "Authorization",
			headerValue:   "Bearer non-existent",
			procedureName: "/bff.v1.ModerationService/Ping",
			want: &authContext{
				DID:   "non-existent",
				Actor: nil,
			},
		},
		{
			name:          "no header",
			procedureName: "/bff.v1.ModerationService/CreateActor",
			wantErr:       "unauthenticated: no token provided",
		},
		{
			name:          "malformed header",
			headerKey:     "Authorization",
			headerValue:   "rewgwegnmwerogkmowergiopwergiopwergop",
			procedureName: "/bff.v1.ModerationService/CreateActor",
			wantErr:       "unauthenticated: malformed header",
		},
		{
			name:          "unsupported auth type",
			headerKey:     "Authorization",
			headerValue:   "OtherType foo",
			procedureName: "/bff.v1.ModerationService/CreateActor",
			wantErr:       "unauthenticated: only Bearer auth supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mag := memoryActorGetter{}
			if tt.actor != nil {
				mag[tt.actor.Did] = tt.actor
			}
			ae := &AuthEngine{
				ActorGetter: mag,
				TokenValidator: func(ctx context.Context, token string) (did string, err error) {
					return token, nil
				},
				Log: zaptest.NewLogger(t),
			}

			req := connect.NewRequest(&v1.PingRequest{})
			if tt.headerKey != "" {
				req.Header().Set(tt.headerKey, tt.headerValue)
			}
			setSpec(req, connect.Spec{Procedure: tt.procedureName})
			got, err := ae.auth(ctx, req)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Empty(t, cmp.Diff(tt.want, got, protocmp.Transform()))
		})
	}
}
