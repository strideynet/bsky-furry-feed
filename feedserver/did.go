package feedserver

import (
	"fmt"
	"net/http"
)

func didHandler(hostname string) (string, http.HandlerFunc) {
	return "/.well-known/did.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{"@context":["https://www.w3.org/ns/did/v1"],"id":"did:web:%s","service":[{"id":"#bsky_fg","type":"BskyFeedGenerator","serviceEndpoint":"https://%s"}]}`, hostname, hostname)))
	}
}
