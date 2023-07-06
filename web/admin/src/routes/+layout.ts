import { get } from 'svelte/store';

import { browser } from '$app/environment';
import { agent, session, setupAgent, setupSession } from '$lib/atp';

import type { LayoutLoad } from './$types';

export const load = (({ url }) => {
  if (browser) {
    if (!get(agent)) {
      console.log('Setting up agent');
      agent.set(setupAgent());
    }

    if (!get(session)) {
      console.log('Setting up session');
      const { session: sessionData, subscriber } = setupSession();

      session.set(sessionData);
      session.subscribe(subscriber);
    }

    const currentSession = get(session);

    if (currentSession) {
      console.log('Resuming session');
      get(agent)?.resumeSession(currentSession);
    }
  }

  return { url };
}) satisfies LayoutLoad;
