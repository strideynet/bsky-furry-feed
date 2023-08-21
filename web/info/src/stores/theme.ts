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

const persistTheme = (t: typeof APP_THEMES[0 | 1]) =>
  (document.cookie = `${APP_THEME_COOKIE_NAME}=${t};path=/;max-age=${60 * 60 * 24}`);

const prefersDark = useMediaQuery('(prefers-color-scheme: dark)'),
  prefersLight = useMediaQuery('(prefers-color-scheme: light)');

if (browser) {
  prefersDark.subscribe(
    (dark) =>
      dark && !document.cookie.includes(APP_THEME_COOKIE_NAME) && theme.set(APP_THEMES[1])
  );
  prefersLight.subscribe(
    (light) =>
      light &&
      !document.cookie.includes(APP_THEME_COOKIE_NAME) &&
      theme.set(APP_THEMES[0])
  );

  theme.subscribe((t) => {
    document.documentElement.classList.remove(...APP_THEMES);
    document.documentElement.classList.add(t);
  });
}

export { persistTheme, theme };
