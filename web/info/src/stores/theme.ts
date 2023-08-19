import { writable } from 'svelte/store';
import { useMediaQuery } from 'svelte-breakpoints';

import { browser } from '$app/environment';
import { APP_THEME_COOKIE_NAME, APP_THEMES } from '$lib/constants';

const initialValue = browser
  ? (document.cookie
      ?.split('; ')
      ?.find((row) => row.startsWith(APP_THEME_COOKIE_NAME))
      ?.split('=')[1] as typeof APP_THEMES[number]) ?? APP_THEMES[0]
  : APP_THEMES[0];

const theme = writable<typeof APP_THEMES[0 | 1]>(initialValue);

const prefersDark = useMediaQuery('(prefers-color-scheme: dark)'),
  prefersLight = useMediaQuery('(prefers-color-scheme: light)');

if (browser) {
  // Only set theme automatically if no cookie is set
  prefersDark.subscribe((dark) => {
    if (dark && !document.cookie.includes(APP_THEME_COOKIE_NAME)) {
      theme.set('dark');
    }
  });
  prefersLight.subscribe((light) => {
    if (light && !document.cookie.includes(APP_THEME_COOKIE_NAME)) {
      theme.set('light');
    }
  });
}

theme.subscribe((t) => {
  if (browser) {
    document.cookie = `${APP_THEME_COOKIE_NAME}=${t};path=/;max-age=${60 * 60 * 24}`;
    document.documentElement.classList.remove(...APP_THEMES);
    document.documentElement.classList.add(t);
  }
});

export { theme };
