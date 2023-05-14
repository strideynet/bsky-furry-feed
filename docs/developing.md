# Developing

Random commands I run often:

```sh
migrate create -ext sql -dir store/migrations -seq create_post_candidate
sqlc generate --experimental
# For local
migrate -path store/migrations -database "postgres://bff:bff@localhost:5432/bff?sslmode=disable" up
# For production with cloud-sql-proxy
migrate -path store/migrations -database "postgres://noah@noahstride.co.uk@localhost:15432/bff?sslmode=disable" up
```

Bluesky client with feed support: https://skyfeed.app/#/

You need HTTPS, so use Cloudflare Tunnel or similar when developing.
