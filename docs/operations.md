# Operations

IaC can be found in the private [infra][infrarepo] repository.

[infrarepo]: https://github.com/furrylist/infra

## Deploying

For now, we’re back to manual deployments:

1. Create a release on GitHub with a new tag.
1. Wait for docker image build to succeed.
1. If migration is necessary, execute it manually via `psql` and bump up the version in the `schema_migrations` table, until we have better tooling.
1. Update the version tag in the [infra repo][infrarepo].
1. On the server, update the `furrylist-infra` repo and run `docker compose up -d`.
1. Celebrate! 🎉
1. If feeds were changed or added since the last deployment, run the **Deploy Feeds** CI job.

## Incident runbooks

Sign in to <https://furrylist.grafana.net> and head to the **Overview**
dashboard. Look at the colorful metrics and hope for improvement!

### Slow generation due to Postgres query planner misbehaving

If the **CPU usage** and **Feed Generation Avg Duration** are unusually
high, this can be due to Postgres misbehaving and incorrectly planning
queries. This can usually be fixed by running the following command in
psql on production:

```sh
bff=# analyze verbose candidate_posts;
INFO:  analyzing "public.candidate_posts"
INFO:  "candidate_posts": scanned 197354 of 197354 pages, containing 2024025 live rows and 2666 dead rows; 300000 rows in sample, 2024025 estimated total rows
ANALYZE
Time: 14113.331 ms (00:14.113)
```
