import { getPageDocument } from '$lib/sanity';

import type { PageLoad } from './$types';

export const load = (async ({ parent, url }) => {
  const preview = url.searchParams.get('preview') === 'true',
    token = url.searchParams.get('token') ?? undefined;

  const { feeds, featuredFeeds } = await parent();

  const content = await getPageDocument('home', preview, token);

  return { content, feeds: { feeds, featured: featuredFeeds } };
}) satisfies PageLoad;
