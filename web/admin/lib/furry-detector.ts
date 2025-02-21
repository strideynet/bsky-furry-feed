import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { matchTerms } from "./util";

type ProfileViewMinimal = Pick<ProfileViewDetailed, "displayName"> &
  Pick<ProfileViewDetailed, "description"> &
  Pick<ProfileViewDetailed, "handle">;

export function isProbablyFurry(profile?: ProfileViewMinimal): boolean {
  if (!profile) {
    return false;
  }

  // ∆ (increment operator) and Δ (delta)
  // Θ (uppercase theta) and θ (lowercase theta)
  const therian = /(Θ|θ)(∆|Δ)/i;

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
    /\b[bp]up(py)?\b/,
    /\bfurs?\b/,
    "anthro",
    "canine",
    "feline",
    /bu?n+u*y/, // too good to not use
    "kemono",
    "furaffinity",
    "derg",
    /scal(y|ie)/,
    /gay (fur|dog|cat|wolf|fox)/,
    /(f|m)urr?suit/,
    /gr(e|a)ymuzzle/,
    /\b(co)?yote\b/,
    "kitsune",
    "hyena",
    /\byeen\b/,
    "otherkin",
    "protogen",
    "fluffy",
    "dog",
    "deer",
    /cat\b/,
    "wolf",
    "dragon",
    /\bsnep\b/,
    "critter",
    "jackalope",
    "tiger",
    "otter",
    "kobold",
    "lion",
    "squirrel",
    /\bpaws?\b/,
    /\bbirb\b/,
    /\b(fur)?sona\b/,
    "cartoon animal",
    "lynx",
    "⨺⃝", // ugly variant of therian theta-delta
    "福瑞", // mandarin for furry
    "babyfur",
    "avali",
    "fennec fox",
    "floofer",
    "dingo",
  ];

  const description = [
    profile.displayName || "",
    profile.handle,
    profile.description,
  ]
    .map((s) => s.toLowerCase())
    .join(" ");

  return matchTerms(terms, description);
}
