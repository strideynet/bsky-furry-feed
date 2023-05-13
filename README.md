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

IaC can be found in `infra/`. The state and deployment is managed by Spacelift.

### Deploying

If database migration is necessary, run this from your workstation before
updating any components.

#### Ingester VM

Container currently managed in GCE directly as lack of support in Terraform for 
GCE container VMs.

Set container metadata to enable logging:

```
google-logging-enabled true
google-logging-use-fluentbit true
```

#### SQL

Terraform/GCP makes it difficult to apply permissions to Cloud SQL IAM users.
Therefore, we must manually grant these accounts permissions:
- Generate a password for the `postgres` built-in user
- Connect to this user.
- Run `apply_permissions.sql`
- Generate a new password for this user, and dispose of it.

Once these are set up, you can open a local proxy with:

```
gcloud config set project bsky-furry-feed
gcloud auth application-default login
./cloud-sql-proxy --auto-iam-authn bsky-furry-feed:us-east1:main-us-east -p 15432
```

When authenticating provide your username and no password.