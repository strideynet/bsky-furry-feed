# bsky-furry-feed

A Bluesky custom feed generator for furry content !

It's also a pretty neat example of a feed generator written in Go.

## Developing

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

## Operations

IaC can be found in `infra/`. The state and deployment is managed by Spacelift.

### Deploying

If database migration is necessary, run this from your workstation before
updating any components.

Update the container image used by the GCE VMs, reboot them.

#### Ingester VM

Container currently managed in GCE directly as lack of support in Terraform for 
GCE container VMs.

Set container metadata to enable logging:

```
google-logging-enabled true
google-logging-use-fluentbit true
```

Standard logging/metrics not ticked.

#### SQL

You can open a local proxy with:

```sh
gcloud config set project bsky-furry-feed
gcloud auth application-default login
# Port 15432 is used to differentiate from the local development postgres
# instance.
./cloud-sql-proxy --auto-iam-authn bsky-furry-feed:us-east1:main-us-east -p 15432
```

When authenticating provide your username/email and no password, IAM auth takes
care of the "password" element (short lived tokens are injected).

Permissions may be a bit screwy at the moment. You'll need to manually grant 
access to tables to service accounts.

At a later date I should:
- Use migrations to create a role.
- Grant this role access to all tables.
- Add guidance on granting this role to cloudsqliamserviceaccount and cloudsqliamuser.