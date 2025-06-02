import { ComAtprotoLabelDefs } from "@atproto/api";
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { newAgent } from "~/lib/auth";
import { getProfile } from "~/lib/cached-bsky";

// A list of labelers that are relevant to and likely trusted by the
// broader Bluesky community and Furrylist user base.
const labelers: Array<Labeler> = [
  {
    did: "did:plc:ar7c4by46qjdydhdevvrndac",
    name: "bsky-mod",
  },
  {
    did: "did:plc:4ugewi6aca52a62u62jccbl7",
    name: "asukafield.xyz",
  },
  {
    did: "did:plc:bv2ckchoc76yobfhkrrie4g6",
    name: "blacksky.app",
  },
  {
    did: "did:plc:lcdcygpdeiittdmdeddxwt4w",
    name: "laelaps.fyi",
  },
];

type Labeler = {
  did: string;
  name: string;
};

export type BlueskyLabel = ComAtprotoLabelDefs.Label & {
  labeler: ProfileViewDetailed;
};

export default async function (did: string): Promise<Array<BlueskyLabel>> {
  const agent = newAgent();
  let cursor: string | undefined = undefined;
  const allLabels: Array<BlueskyLabel> = [];
  do {
    const labels = await agent.com.atproto.label.queryLabels({
      uriPatterns: [did],
      sources: labelers.map((l) => l.did),
      cursor,
    });
    cursor = labels.data.cursor;
    for (const label of labels.data.labels) {
      allLabels.push({
        ...label,
        labeler: await getProfile(label.src),
      });
    }
  } while (cursor);
  return allLabels;
}
