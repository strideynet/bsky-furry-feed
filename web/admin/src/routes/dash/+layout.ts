import { get } from 'svelte/store';
import { redirect } from '@sveltejs/kit';

import { agent, session } from '$lib/atp';

import type { LayoutLoad } from './$types';

export const ssr = false;

export const load = (async ({ parent }) => {
  await parent();

  if (!get(session) || !get(agent)?.hasSession) {
    throw redirect(302, '/auth/login');
  }
}) satisfies LayoutLoad;
