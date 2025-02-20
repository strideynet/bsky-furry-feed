import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";

export type ProfileViewMinimal = Pick<ProfileViewDetailed, "displayName"> &
  Pick<ProfileViewDetailed, "description"> &
  Pick<ProfileViewDetailed, "handle"> &
  Pick<ProfileViewDetailed, "followsCount">;

export function addSISuffix(number?: number) {
  number = number || 0;

  const suffixes = ["", "K", "M"];
  const order = Math.floor(Math.log10(number) / 3);

  for (let i = 0; i < order; i++) {
    number = number / 1000;
  }

  return `${Math.round(number * 100) / 100}${suffixes[order] || ""}`;
}

export function matchTerms(
  terms: (string | RegExp)[],
  haystack: string
): boolean {
  for (const term of terms) {
    if (
      typeof term === "object" ? haystack.match(term) : haystack.includes(term)
    ) {
      return true;
    }
  }

  return false;
}
