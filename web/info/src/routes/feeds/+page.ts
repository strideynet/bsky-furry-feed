import type { PageLoad } from './$types';

export const load = (async ({ parent }) => {
  const { feeds } = await parent();

  return { feeds };
}) satisfies PageLoad;
