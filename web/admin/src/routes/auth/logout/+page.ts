import { get } from 'svelte/store';
import { redirect } from '@sveltejs/kit';

import { session } from '$lib/atp';
import { ATP_SESSION_COOKIE } from '$lib/constants';

import type { PageLoad } from './$types';

export const ssr = false;

export const load = (async ({ parent }) => {
  await parent();

  if (!get(session)) {
    throw redirect(302, '/');
  }

  session.set(null);
  localStorage.removeItem(ATP_SESSION_COOKIE);

  throw redirect(302, '/');
}) satisfies PageLoad;
