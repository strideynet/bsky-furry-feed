import { browser } from '$app/environment';
import { writable } from 'svelte/store';

import { APP_THEMES } from '$lib/constants';

let themeValue: number = APP_THEMES.light;

const themeWritable = writable<typeof APP_THEMES[keyof typeof APP_THEMES]>(themeValue);

// TODO: Clean up this spagetti

if (browser) {
  const localStorageTheme = localStorage.getItem('theme');

  if (localStorageTheme) {
    themeWritable.set(parseInt(localStorageTheme));
  } else {
    themeWritable.set(
      window.matchMedia('(prefers-color-scheme: dark)')
        ? APP_THEMES.dark
        : APP_THEMES.light
    );
  }
}

const updateTheme = (value: number) => {
  if (browser) {
    localStorage.setItem('theme', value.toString());

    if (value == APP_THEMES.dark) {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    } else {
      document.documentElement.classList.add('light');
      document.documentElement.classList.remove('dark');
    }
  }
};

// TODO: Remove this, temp workaround
themeWritable.subscribe((value) => {
  themeValue = value;

  if (browser) {
    updateTheme(value);
  }
});

if (browser) {
  updateTheme(themeValue);
}

export default themeWritable;
