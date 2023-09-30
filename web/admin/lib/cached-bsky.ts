import { newAgent } from "./auth";
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";

class BatchDidQueue<K, T extends { did: string }> {
  private items: { key: K; resolve: (value: T | PromiseLike<T>) => void }[] =
    [];
  private timer: any = null;
  private readonly maxItems = 25;

  constructor(private readonly exec: (k: K[]) => Promise<T[]>) {}

  async add(key: K): Promise<T> {
    return new Promise((resolve) => {
      this.items.push({ key, resolve });

      if (this.items.length >= this.maxItems) {
        this.dispatch();
      } else if (!this.timer) {
        this.timer = setTimeout(() => this.dispatch(), 50);
      }
    });
  }

  private async dispatch() {
    if (this.items.length === 0) return;
    const keys = this.items.slice(0, this.maxItems);
    this.items = this.items.slice(this.maxItems);
    this.timer = null;

    const results = await this.exec(keys.map((k) => k.key));

    for (const { key, resolve } of keys) {
      const result = results.find((result) => result.did === key);
      resolve(result as T);
    }
  }
}

const cache: Map<string, Promise<ProfileViewDetailed>> = new Map();
const queue = new BatchDidQueue((dids: string[]) =>
  newAgent()
    .getProfiles({ actors: dids })
    .then((r) => r.data.profiles)
);

export async function getProfile(did: string): Promise<ProfileViewDetailed> {
  let profile = cache.get(did);
  if (profile) {
    return profile;
  }

  profile = queue.add(did);

  cache.set(did, profile);

  return profile;
}
