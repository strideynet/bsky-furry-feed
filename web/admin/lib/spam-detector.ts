import { matchTerms, ProfileViewMinimal } from "./util";

const FOLLOWS_THRESHOLD = 10_000;

export function isProbablySpam(profile?: ProfileViewMinimal): boolean {
  if (!profile) {
    return false;
  }

  if ((profile.followsCount || 0) > FOLLOWS_THRESHOLD) {
    return true;
  }

  if (!profile.description) {
    return false;
  }

  const terms = [
    /#resist(er)?\b/i,
    /#teamblue\b/i,
    /#bluecrew\b/i,
    /\bai artist\b/i,
    /blue (democrat|crew)/i,
    /#defenddemocracy\b/i,
    /\b(ai |to )?prompt\b/i,
    /(dm|message|e?mail)( me)? (for|to) (removal|remove)/i,
    /follow\b.+follow back/,
  ];

  return matchTerms(terms, profile.description);
}
