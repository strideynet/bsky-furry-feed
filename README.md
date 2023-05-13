# bsky-furry-feed

A Bluesky custom feed generator for furry content !

It's also a pretty neat example of a feed generator written in Go.

## Developing

Random commands I run often:

```sh
migrate create -ext sql -dir store/migrations -seq create_post_candidate
sqlc generate --experimental
migrate -path store/migrations -database "postgres://bff:bff@localhost:5432/bff?sslmode=disable" up
```

Bluesky client with feed support: https://skyfeed.app/#/

You need HTTPS, so use Cloudflare Tunnel or similar when developing.

## Operations

IaC can be found in `infra/`.

### Deploying

If database migration is necessary, run this from your workstation.

Manually bump container image versions in `infra/`.

