import { writable } from 'svelte/store';

import { ATP_API, ATP_SESSION_COOKIE } from '$lib/constants';

import * as atproto from '@atproto/api';

import type { AtpSessionData, AtpSessionEvent } from '@atproto/api';

const session = writable<AtpSessionData | null>(null),
  agent = writable<atproto.BskyAgent | null>(null);

const setupSession = () => {
  let session: AtpSessionData | null = null;

  const sessionCookie = localStorage.getItem(ATP_SESSION_COOKIE) || null;

  const subscriber = (data: typeof session) => {
    if (data === null) {
      localStorage.removeItem(ATP_SESSION_COOKIE);
      return;
    }

    try {
      const currentData = JSON.parse(
        localStorage.getItem(ATP_SESSION_COOKIE) || '{}'
      ) as AtpSessionData;

      if (currentData !== data) {
        localStorage.setItem(ATP_SESSION_COOKIE, JSON.stringify(data));
      }
    } catch {
      localStorage.removeItem(ATP_SESSION_COOKIE);
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

export { agent, session, setupAgent, setupSession };
