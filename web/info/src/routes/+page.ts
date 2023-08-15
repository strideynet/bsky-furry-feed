import type { PageLoad } from './$types';

export const load = (async ({ parent, fetch: _fetch }) => {
  const { feeds, featuredFeeds } = await parent();

  return { feeds, featuredFeeds };
}) satisfies PageLoad;
