import { get } from 'svelte/store';
import { redirect } from '@sveltejs/kit';

import { agent, session, setupAgent, setupSession } from '$lib/atp';

import type { LayoutLoad } from './$types';

export const ssr = false;

export const load = (({ url }) => {
  if (!get(agent)) {
    agent.set(setupAgent());
  }

  if (!get(session)) {
    const { session: sessionData, subscriber } = setupSession();

    if (sessionData) {
      session.set(sessionData);
    }
    session.subscribe(subscriber);
  }

  const currentSession = get(session);

  if (currentSession && !get(agent)?.hasSession) {
    get(agent)?.resumeSession(currentSession);
  }

  if (!get(session) && !url.pathname.startsWith('/auth')) {
    throw redirect(302, '/auth/login');
  }

  return { url };
}) satisfies LayoutLoad;
