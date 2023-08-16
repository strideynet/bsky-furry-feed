import { error } from '@sveltejs/kit';

import { DOC_ROUTES_NAMES } from '$lib/constants';
import { getPageDocument } from '$lib/sanity';

import type { PageLoad } from './$types';

export const load = (async ({ parent, params }) => {
  // @ts-expect-error This is fine in this ctx
  const pageName = DOC_ROUTES_NAMES[params.page] as keyof typeof DOC_ROUTES_NAMES;

  if (!pageName) {
    throw error(404, 'Not Found');
  }

  const { feeds, featuredFeeds } = await parent();

  const content = getPageDocument(pageName);

  return { content, feeds: { feeds, featured: featuredFeeds } };
}) satisfies PageLoad;
