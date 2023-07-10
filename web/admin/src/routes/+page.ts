import { redirect } from '@sveltejs/kit';

import type { PageLoad } from './$types';

export const ssr = false;

export const load = (async ({ parent }) => {
  await parent();

  // This route is just a redirect to the dashboard when authenticated.
  throw redirect(302, '/dash');
}) satisfies PageLoad;
