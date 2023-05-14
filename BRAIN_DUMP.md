# Planning

## Candidates

We refer to things used in the process of building the feed as "candidates".

A candidate "repository" is a individual user/account on BlueSky that's identified
by a DID. Candidate repositories will be a static list of repositories.

Build order:

1. Users manually opt in to the service via contact with Noah
2. User portal to allow users to opt their repositories in
3. Moderation portal to allow moderators to opt repositories in and out
4. Determine "candidate" users for moderators to approve/disapprove.
    - Find users who are often interacted with by furries
    - Apply some rough detection to profile description and create a "score"
    - Moderators can then approve/disapprove

We will attach some metadata 

We filter repository updates down to those coming from candidate repositories.

For now, we will store:

- Posts (including replies, which should be distinctly marked)
- Reposts
- Likes

In future, we may also store Follows in order to "detect" other furries.

## Feeds

Planned feeds in order of implementation:
- Furry chronological
- Furry "whats hot"

Eventually allow these feeds to be broken down by AD and Artist status.

# Bluesky schema dumps

## `app.bsky.feed.repost`

```yaml
# `at://did:plc:dllwm3fafh66ktjofzxhylwk/app.bsky.feed.repost/3jvpgz5ik4p26`
 {
  "$type": "app.bsky.feed.repost",
  "createdAt": "2023-05-14T18:09:45.052Z",
  "subject": {
    "cid": "bafyreieflx5rcjwadinfayfnsgzn4cqzj3lncfslam4dysr7ms4xa2xmna",
    "uri": "at://did:plc:ouytv644apqbu2pm7fnp7qrj/app.bsky.feed.post/3jvkfw2ovbc2b"
  }
}
```