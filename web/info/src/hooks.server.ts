import { APP_THEME_COOKIE_NAME, APP_THEMES } from '$lib/constants';

import type { Handle, ResolveOptions } from '@sveltejs/kit';

export const handle = (async ({ event, resolve }) => {
  const resolveOptions: ResolveOptions = {};

  resolveOptions.filterSerializedResponseHeaders = (name: string, _value: string) => {
    switch (name) {
      case 'content-type':
        return true;
      default:
        return false;
    }
  };

  const theme = event.cookies.get(APP_THEME_COOKIE_NAME);

  // @ts-expect-error theme may not be 'light'|'dark'
  if (theme && APP_THEMES.includes(theme)) {
    resolveOptions.transformPageChunk = ({ html }) => {
      const classAttrRegexp = /<html([^>]*?)class="([^"]*?)"/i,
        htmlTagRegexp = /<html([^>]*?)>/i;

      if (classAttrRegexp.test(html)) {
        // if class attribute exists, append the theme to it
        return html.replace(classAttrRegexp, `<html$1class="$2 ${theme}"`);
      } else if (htmlTagRegexp.test(html)) {
        // if <html> tag exists but no class attribute, add class with the theme
        return html.replace(htmlTagRegexp, `<html$1 class="${theme}">`);
      }

      return html;
    };
  }

  return await resolve(event, resolveOptions);
}) satisfies Handle;
