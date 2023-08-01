import { AppBskyActorGetProfile } from "@atproto/api";
import { newAgent } from "./auth";

const cache: Map<string, Promise<AppBskyActorGetProfile.Response>> = new Map();

export async function getProfile(
  did: string
): Promise<AppBskyActorGetProfile.Response> {
  let profile = cache.get(did);
  if (profile) {
    return profile;
  }

  profile = newAgent().getProfile({ actor: did });

  cache.set(did, profile);

  return profile;
}
