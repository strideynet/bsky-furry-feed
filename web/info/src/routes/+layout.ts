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
import type { FeedInfo } from '$types';

const mockFeedData = [
  {
    id: 'furry-new',
    displayName: '🐾 New',
    description: 'Posts by all furries on furryli.st, sorted chronologically.',
    priority: 101,
    link: 'https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-new'
  },
  {
    id: 'furry-nsfw',
    displayName: '🐾 New 🌙',
    description:
      'All posts by furries on furryli.st that have the #nsfw hashtag, sorted chronologically.',
    priority: 100,
    link: 'https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-nsfw'
  },
  {
    id: 'furry-hot',
    displayName: '🐾 Hot',
    description: 'Posts by all furries on furryli.st, sorted by "hotness".',
    priority: 99,
    link: 'https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-hot'
  },
  {
    id: 'furry-fursuit',
    displayName: '🐾 Fursuit',
    description:
      'All posts by furries on furryli.st that have the #fursuit hashtag, sorted chronologically.',
    priority: 98,
    link: 'https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-fursuit'
  },
  {
    id: 'furry-art',
    displayName: '🐾 Art',
    description:
      'All posts by furries on furryli.st that have the #art or #furryart hashtag, sorted chronologically.',
    priority: 97,
    link: 'https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-art'
  },
  {
    id: 'furry-comms',
    displayName: '🐾 #commsopen',
    description:
      'All posts by furries on furryli.st that have the #commsopen hashtag, sorted chronologically.',
    priority: 96,
    link: 'https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-comms'
  }
] satisfies FeedInfo[];

export const load = (async ({ url }) => {
  const feeds = mockFeedData
      // Exclude negative priorities
      ?.filter((feed) => feed.priority >= 0)
      // Sort by priority (descending)
      ?.sort((a, b) => b.priority - a.priority),
    featuredFeeds = feeds?.filter((feed) => feed.priority >= 100) ?? [];

  if (!browser) {
    return { url, feeds, featuredFeeds };
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

  return { url, feeds, featuredFeeds };
}) satisfies LayoutLoad;
