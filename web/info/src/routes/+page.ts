import { getPageDocument } from '$lib/sanity';

import type { PageLoad } from './$types';

export const load = (async ({ parent }) => {
  const { feeds, featuredFeeds } = await parent();

  const content = await getPageDocument('home');

  return { content, feeds: { feeds, featured: featuredFeeds } };
}) satisfies PageLoad;
