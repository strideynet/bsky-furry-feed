export const APP_THEMES = {
  light: 0,
  dark: 1
};

export const BASE_TRANSITION_DURATION = 150;

export const ATP_API = 'https://bsky.social';

export const API_URL = import.meta.env.ADMIN_API_URL || 'https://feed.furryli.st';

export const LOCALSTORAGE_ATP_SESSION_KEY = 'bff-atp-session';

export const ATP_SESSION_COOKIE = 'bff-atp-session';

export const NAV_OPTIONS = [
  {
    href: '/welcome',
    text: 'Welcome'
  },
  {
    href: '/feeds',
    text: 'Feeds'
  },
  {
    href: '/community-guidelines',
    text: 'Community Guidelines'
  },
  {
    href: 'https://discord.gg/7X467r4UXF',
    target: '_blank',
    text: 'Discord'
  }
];

export const VALID_DOC_ROUTES = ['welcome', 'community-guidelines', 'feeds'];

export const DOC_ROUTES_NAMES = {
  welcome: 'welcome',
  'community-guidelines': 'communityGuidelines',
  feeds: 'feeds'
} as const;

export const SANITY_PROJECT_ID = '0ildj6pc' as const;

export const SANITY_DATASET = 'production';

export const SANITY_API_VERSION = '2022-11-29';
