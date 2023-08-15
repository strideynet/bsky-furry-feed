import { get } from 'svelte/store';

import { browser } from '$app/environment';
import { getClient } from '$lib/api';
import {
  agent,
  fetchProfile,
  profile,
  session,
  setupAgent,
  setupSession
} from '$lib/atp';

import type { LayoutLoad } from './$types';
import type { FeedInfo } from '$types';

let feeds: FeedInfo[] | null = null,
  featuredFeeds: FeedInfo[] | null = null;

export const load = (async ({ url, fetch }) => {
  const apiClient = getClient(fetch);

  (feeds ||=
    (await apiClient
      .listFeeds({})
      .then((res) => {
        return res.feeds
          .filter((f) => f.priority >= 0)
          .sort((a, b) => {
            if (a.priority === b.priority) {
              return a.id.localeCompare(b.id);
            }
            return b.priority - a.priority;
          });
      })
      .catch(console.error)) ?? null),
    (featuredFeeds ||= feeds?.filter((feed) => feed.priority >= 100) ?? null);

  if (!browser) {
    return { apiClient, url, feeds, featuredFeeds };
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

  return { apiClient, url, feeds, featuredFeeds };
}) satisfies LayoutLoad;
