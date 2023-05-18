package feedserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
)

func getCandidateActorHandler(
	log *zap.Logger, queries *store.Queries,
) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		handle := r.URL.Query().Get("handle")
		if handle == "" {
			handleErr(w, log, fmt.Errorf("no handle provided"))
			return
		}
		client := bluesky.NewUnauthClient()
		did, err := client.ResolveHandle(r.Context(), handle)
		if err != nil {
			handleErr(w, log, fmt.Errorf("resolving handle: %w", err))
			return
		}
		data, err := queries.GetCandidateActorByDID(r.Context(), did.Did)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				handleErr(w, log, fmt.Errorf("no results found for handle"))
				return
			}
			handleErr(w, log, fmt.Errorf("getting candidate actor: %w", err))
			return
		}
		candidateRepository := bff.CandidateActorFromStore(data)
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", " ")
		_ = encoder.Encode(candidateRepository)
	}
	return "/get_candidate_actor", otelhttp.NewHandler(h, "get_candidate_actor")
}
