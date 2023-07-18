import { get } from 'svelte/store';
import { redirect } from '@sveltejs/kit';

import { session } from '$lib/atp';

import type { PageLoad } from './$types';

export const ssr = false;

export const load = (async ({ parent }) => {
  await parent();

  if (!get(session)) {
    throw redirect(302, '/');
  }
}) satisfies PageLoad;
