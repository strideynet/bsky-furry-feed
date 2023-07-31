import { get } from 'svelte/store';

import { browser } from '$app/environment';
import {
  agent,
  fetchProfile,
  profile,
  session,
  setupAgent,
  setupSession
} from '$lib/atp';

import type { LayoutLoad } from './$types';

export const load = (async ({ url }) => {
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

  const currentSession = get(session),
    currentAgent = get(agent);

  if (currentSession && currentAgent && !currentAgent.hasSession) {
    if (!currentAgent.hasSession) {
      await currentAgent.resumeSession(currentSession);
    }
    if (!get(profile)) {
      await fetchProfile(currentAgent, currentSession);
    }
  }

  return { url };
}) satisfies LayoutLoad;
