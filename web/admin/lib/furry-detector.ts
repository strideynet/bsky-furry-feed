import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";

type ProfileViewMinimal = Pick<ProfileViewDetailed, "displayName"> &
  Pick<ProfileViewDetailed, "description">;

export function isProbablyFurry(profile?: ProfileViewMinimal): boolean {
  if (!profile) {
    return false;
  }

  // ∆ (increment operator) and Δ (delta)
  // Θ (uppercase theta) and θ (lowercase theta)
  const therian = /(Θ|θ)(∆|Δ)/;

  if (profile?.displayName?.match(therian)) {
    return true;
  }

  if (!profile.description) {
    return false;
  }

  const terms = [
    "furry",
    "furries",
    therian,
    "therian",
    /\bpup\b/,
    /\bfur\b/,
    "anthro",
    "canine",
    /bu?n+u*y/, // too good to not use
    "kemono",
    "furaffinity",
    "derg",
    /scal(y|ie)/,
    /gay (fur|dog|cat|wolf)/,
    /(f|m)urr?suit/,
    "otherkin",
    "protogen",
    "fluffy",
  ];

  const description = profile.description.toLowerCase();

  for (const term of terms) {
    if (
      typeof term === "object"
        ? description.match(term)
        : description.includes(term)
    ) {
      return true;
    }
  }

  return false;
}