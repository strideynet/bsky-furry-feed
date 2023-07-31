import { get, writable } from 'svelte/store';

import { ATP_API, ATP_SESSION_COOKIE } from '$lib/constants';

import * as atproto from '@atproto/api';

import type {
  AppBskyActorGetProfile,
  AtpSessionData,
  AtpSessionEvent
} from '@atproto/api';

const session = writable<AtpSessionData | null>(null),
  agent = writable<atproto.BskyAgent | null>(null),
  profile = writable<AppBskyActorGetProfile.Response['data'] | null>(null);

const fetchProfile = async (agent: atproto.BskyAgent | null, session: AtpSessionData) => {
  if (!agent) {
    return;
  }

  const response = await agent.getProfile({ actor: session.did });
  if (response.success === false) {
    return;
  }

  profile.set(response.data);
};

const setupSession = () => {
  let session: AtpSessionData | null = null;

  const sessionCookie = localStorage.getItem(ATP_SESSION_COOKIE) || null;

  const subscriber = (data: typeof session) => {
    if (data === null) {
      localStorage.removeItem(ATP_SESSION_COOKIE);
      profile.set(null);
      return;
    }

    try {
      const currentData = JSON.parse(
        localStorage.getItem(ATP_SESSION_COOKIE) || '{}'
      ) as AtpSessionData;

      if (currentData !== data) {
        localStorage.setItem(ATP_SESSION_COOKIE, JSON.stringify(data));
        fetchProfile(get(agent), data);
      }
    } catch {
      localStorage.removeItem(ATP_SESSION_COOKIE);
      profile.set(null);
    }
  };

  if (!sessionCookie) {
    return { session, subscriber };
  }

  try {
    const sessionData = JSON.parse(sessionCookie) as AtpSessionData;

    if (sessionData) {
      session = sessionData;
    }

    return { session, subscriber };
  } catch {
    return { session, subscriber };
  }
};

const setupAgent = () => {
  const persistSessionWith = (e: AtpSessionEvent, data?: AtpSessionData) => {
    switch (e) {
      case 'update':
      case 'create':
        if (!data) {
          console.error('No session data was provided');
          break;
        }
        break;
      case 'expired':
        session.set(null);
        break;
      case 'create-failed':
        session.set(null);
        console.error('Failed to create session');
        break;
    }
  };

  return new atproto.BskyAgent({ service: ATP_API, persistSession: persistSessionWith });
};

export { agent, fetchProfile, profile, session, setupAgent, setupSession };
