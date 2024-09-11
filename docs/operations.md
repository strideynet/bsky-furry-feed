# Operations

IaC can be found in `infra/`. The state and deployment is managed by Spacelift.

## Deploying

If database migration is necessary, run this from your workstation before
updating any components.

Update the images in `infra/k8s` and `kubectl apply`.

### CloudSQL

You can open a local proxy to the production database with:

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
access to tables to cloudsqliamserviceaccount and cloudsqliamuser.

### Runbook

1. Cut release on GitHub with tag
2. Wait for docker image build to succeed
3. If migration is necessary. If so, run it

```sh
$ migrate -path store/migrations -database "postgres://noah@noahstride.co.uk@localhost:15432/bff?sslmode=disable" up
```

4. If migration created tables, assign permissions to the correct groups
5. Update k8s manifests
6. Apply manifests
7. Monitor deployment
8. Celebrate