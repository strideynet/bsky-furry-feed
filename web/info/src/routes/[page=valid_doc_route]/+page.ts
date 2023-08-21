import { error } from '@sveltejs/kit';

import { DOC_ROUTES_NAMES } from '$lib/constants';
import { getPageDocument } from '$lib/sanity';

import type { PageLoad } from './$types';

export const load = (async ({ parent, params, url }) => {
  const preview = url.searchParams.get('preview') === 'true',
    token = url.searchParams.get('token') ?? undefined;

  // @ts-expect-error This is fine in this ctx
  const pageName = DOC_ROUTES_NAMES[
    params.page
  ] as typeof DOC_ROUTES_NAMES[keyof typeof DOC_ROUTES_NAMES];

  if (!pageName) {
    throw error(404, 'Not Found');
  }

  const { feeds, featuredFeeds } = await parent();

  const content = getPageDocument(pageName, preview, token);

  return { content, feeds: { feeds, featured: featuredFeeds } };
}) satisfies PageLoad;
