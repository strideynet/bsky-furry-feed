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

  if (theme && APP_THEMES.includes(theme)) {
    resolveOptions.transformPageChunk = ({ html }) => {
      const match = html.match(/(<html.*?)(>)/);

      if (!match) {
        return html;
      }

      const startHtml = match[1],
        endTag = match[2],
        classes = startHtml.includes('class="') ? ' ' : ' class="',
        newHtml = `${startHtml}${classes}${theme}"${endTag}`;

      return html.replace(match[0], newHtml);
    };
  }

  return await resolve(event, resolveOptions);
}) satisfies Handle;
