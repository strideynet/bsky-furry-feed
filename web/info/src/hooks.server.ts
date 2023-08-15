import type { Handle } from '@sveltejs/kit';

export const handle = (async ({ event, resolve }) => {
  const response = await resolve(event, {
    // As per https://kit.svelte.dev/docs/hooks#server-hooks-handle, we need
    // to specify headers to be included in serialized responses when a `load` function
    // loads a resource server-side with a `RouteFetch`
    filterSerializedResponseHeaders: (name: string, _value: string) => {
      switch (name) {
        case 'content-type':
          return true;
        default:
          return false;
      }
    }
  });

  return response;
}) satisfies Handle;
