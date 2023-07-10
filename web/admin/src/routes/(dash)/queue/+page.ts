import type { PageLoad } from './$types';

export const ssr = false;

export const load = (async ({ parent, fetch }) => {
  await parent();

  // TODO: Implement fetching of queue item(s).

  return { queue: [], routeFetch: fetch };
}) satisfies PageLoad;
