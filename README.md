# bsky-furry-feed

A Bluesky custom feed generator for furry content !

It's also a pretty neat example of a feed generator written in Go.

## Useful commands for hacking

```sh
migrate create -ext sql -dir store/migrations -seq create_post_candidate
sqlc generate --experimental
migrate -path store/migrations -database "postgres://bff:bff@localhost:5432/bff?sslmode=disable" up
```