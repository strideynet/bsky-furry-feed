import type { PageLoad } from './$types';

export const load = (async ({ parent, fetch: _fetch }) => {
  const parentData = await parent();

  const feeds = parentData.feeds
    // Exclude negative priorities
    ?.filter((feed) => feed.priority >= 0)
    // Sort by priority (descending)
    ?.sort((a, b) => b.priority - a.priority);

  const featuredFeeds = feeds?.filter((feed) => feed.priority >= 100);

  return { feeds, featuredFeeds };
}) satisfies PageLoad;
