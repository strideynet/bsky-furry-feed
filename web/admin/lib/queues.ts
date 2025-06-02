import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { isProbablySpam } from "./spam-detector";
import { isProbablyFurry } from "./furry-detector";
import { Actor } from "../../proto/bff/v1/types_pb";
import { UnwrapRef } from "nuxt/dist/app/compat/capi";

type UnwrappedActor = UnwrapRef<Actor>;

export const queueTypes = [
  "All",
  "Likely furry",
  "Likely spam",
  "Empty",
  "Held back",
] as const;

const includeInAll: Array<Category> = ["Likely furry"];

type Category = typeof queueTypes[number];

function categorizeProfile(
  actor: UnwrappedActor,
  profile?: ProfileViewDetailed
): Category {
  if (actor.heldUntil && actor.heldUntil.toDate() > new Date())
    return "Held back";
  if (isProbablySpam(profile)) return "Likely spam";
  if (isProbablyFurry(profile)) return "Likely furry";
  if (!profile) return "Empty";
  if (!profile.displayName && !profile.description && !profile.postsCount)
    return "Empty";

  return "All";
}

export function categorizeProfiles(
  actors: Array<UnwrappedActor>,
  profiles: Map<string, ProfileViewDetailed>
): Record<Category, Array<UnwrappedActor>> {
  const result = {} as Record<Category, Array<UnwrappedActor>>;

  for (const type of queueTypes) {
    result[type] = [];
  }

  // Categorize each actor and add to appropriate category
  for (const actor of actors) {
    const profile = profiles.get(actor.did);
    const category = categorizeProfile(actor, profile);

    result[category].push(actor);

    if (includeInAll.includes(category)) {
      result["All"] = result["All"] || [];
      result["All"].push(actor);
    }
  }

  return result;
}
