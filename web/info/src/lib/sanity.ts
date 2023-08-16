import { SANITY_API_VERSION, SANITY_DATASET, SANITY_PROJECT_ID } from '$lib/constants';

import { createClient } from '@sanity/client';

import type { DocumentRegistry } from '$types';

const client = createClient({
  projectId: SANITY_PROJECT_ID,
  dataset: SANITY_DATASET,
  useCdn: true,
  apiVersion: SANITY_API_VERSION
});

export const getPageDocument = async <T extends string>(id: T) => {
  const document = await client
    .fetch<T extends keyof DocumentRegistry ? DocumentRegistry[T] : never>(
      `*[_type == "page" && id == "${id}" && !(_id in path('drafts.**'))][0]`
    )
    .then((response) => {
      const document = response;
      return document;
    })
    .catch((err) => {
      console.error(err);
      return undefined;
    });

  return document;
};
