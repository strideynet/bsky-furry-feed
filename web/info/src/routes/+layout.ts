import { get } from 'svelte/store';

import { browser } from '$app/environment';
import { agent, session, setupAgent, setupSession } from '$lib/atp';

import type { LayoutLoad } from './$types';

export const load = (({ url }) => {
  if (!browser) {
    return { url };
  }

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

  if (currentSession) {
    get(agent)?.resumeSession(currentSession);
  }

  return { url };
}) satisfies LayoutLoad;
