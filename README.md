# bsky-furry-feed

The source code and infrastructure for `https://feed.furryli.st`.

It produces a custom feed for the Bluesky social media site, selecting posts
based on people's membership of the furry community!

## For furries

- Open https://skyfeed.app/, a third-party client
- Login with an app password from [bsky settings](https://staging.bsky.app/settings/app-passwords)
- Under `Custom Feeds`, add one with did: `did:web:feed.furryli.st` and Feed ID: `furry-new`

Reach out on the [Bluesky furries discord](https://discord.gg/5UNyBtnwKy) for more information!

## For developers

This is also a neat example of a Bluesky feed generator written in Go! If you
are trying to build something similar in Go, and need any advice, please learn
what you can from the source code and ask any questions.
