# Developing

The Bluesky furry feed, sometimes referred to as **bff**, is written
mostly in [Go][go]. It broadly consists of a data ingester, the feed
generation service, and a moderation system for—among other
uses—approving new users to the feed.

[go]: https://go.dev

## Getting started

After [cloning this repository][clone], install the following required
system dependencies if you don’t have them yet:

1. The latest [Go 1.20][go] version,
2. [Docker][docker], and
3. [docker-compose][docker-compose] if your Docker version doesn’t
   support running `docker compose` as subcommand of Docker.

In addition to this, we’re using `sqlc` and `migrate` to manage our
database schema. Install them by executing the following commands:

```bash
$ go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.19.0
$ go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

In order to run the ingest and feed server, you need to start a local
development database. Start it by executing `docker-compose up -d`. If
you want to turn it off again, execute `docker-compose down`. You can
learn more about docker-compose [in the official docs][docker-compose].

Now that your database is running, you need to initially run all database
migrations, so that all required tables are created. You’ll need to
re-run this any time we create a new migration.

```sh
$ migrate -path store/migrations -database "postgres://bff:bff@localhost:5432/bff?sslmode=disable" up
1/u candidate_actors (Xms)
2/u create_candidate_post (Xms)
3/u create_candidate_like (Xms)
4/u create_candidate_follow (Xms)
...
```

For migrating the production database with `cloud-sql-proxy`, you (i.e.
probably Noah) run the following command:

```sh
$ migrate -path store/migrations -database "postgres://noah@noahstride.co.uk@localhost:15432/bff?sslmode=disable" up
...
```

[clone]: https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository
[docker]: https://docs.docker.com/engine/install/#server
[docker-compose]: https://docs.docker.com/compose/

## Configuring your environment

Both the server and cli need a Bluesky login to start. Copy the
`.env.example` to `.env` and follow the instructions in there.

## Running the ingest and feed server

With the database running and the schema migrated, you can start the main
bff server (aka. `bffsrv`):

```sh
$ go run ./cmd/bffsrv/
...
```

You may also want to use the cli (aka. `bffctl`) to e.g. interact with
the database or find a did by a user’s handle.

```sh
$ go run ./cmd/bffctl/ -e local
NAME:
   bffctl - The swiss army knife of any BFF operator
...
```

By default, the ingest saves no data because no user is registered as
so-called _candidate actor_. To add a user as candidate actor and allow
the ingest server to collect their posts, likes, and follows, add them
to the database using bffctl (where `HANDLE` is your Bluesky handle, such
as `ottr.sh`):

```sh
$ go run ./cmd/bffctl/ -e local db ca add --handle HANDLE
...
2023-07-07T01:26:56.071+0200    INFO    bffctl/db.go:175        successfully added
```

## Migrations and queries

We use `sqlc` to generate type-safe code from raw SQL queries and
`migrate` to create & manage migrations.

To create a migration, run this command in the project root where
`NAME` is the name of your migration, such as `create_post_candidate`:

```sh
$ migrate create -ext sql -dir store/migrations -seq NAME
/.../store/migrations/000010_NAME.up.sql
/.../store/migrations/000010_NAME.down.sql
```

The queries in the `*.up.sql` file are executed when running the
[`up`](#getting-started) command using migrate. To rollback the latest
migration, run migrate’s `down` command.

```sh
$ migrate -path store/migrations -database "postgres://bff:bff@localhost:5432/bff?sslmode=disable" down 1
10/d NAME (39.10048ms)
```

After applying a new migration or editing a query, such as in
`store/queries/candidate_posts.sql`, we need to generate the sqlc
bindings for the database schema and all queries:

```sh
$ sqlc generate --experimental
```

## Skyfeed

While the official Bluesky client supports feeds, you may prefer using
[SkyFeed][skyfeed], a community-created client with feed support.

You need HTTPS, so use Cloudflare Tunnel or similar when developing.

[skyfeed]: https://skyfeed.app/#/

## Archictural Overview

![image](https://github.com/strideynet/bsky-furry-feed/assets/16336790/14e85bd6-de4f-4bbb-96aa-c6bb4cfc5394)

