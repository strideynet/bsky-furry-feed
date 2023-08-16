import { SANITY_API_VERSION, SANITY_DATASET, SANITY_PROJECT_ID } from '$lib/constants';

import { createClient } from '@sanity/client';
import imageUrlBuilder from '@sanity/image-url';

import type {
  SanityImageObject,
  SanityImageSource
} from '@sanity/image-url/lib/types/types';
import type { DocumentRegistry } from '$types';

const client = createClient({
  projectId: SANITY_PROJECT_ID,
  dataset: SANITY_DATASET,
  useCdn: true,
  apiVersion: SANITY_API_VERSION
});

export const getPageDocument = async <T extends string>(
  id: T,
  preview = false,
  token?: string
) => {
  const localClient = token ? createClient({ ...client.config(), token }) : client;
  const query = `*[_type == "page" && id == "${id}"${
    preview ? '' : ' && !(_id in path("drafts.**"))'
  }][0]`;
  const document = await localClient
    .fetch<T extends keyof DocumentRegistry ? DocumentRegistry[T] : never>(query)
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

export interface ImageCrop {
  top: number;
  left: number;
  bottom: number;
  right: number;
  width: number;
  height: number;
}

const builder = imageUrlBuilder(client);

export const urlFor = (source: SanityImageSource) => builder.image(source);

export const getCrop = (image: SanityImageObject | undefined) => {
  if (!image || !image?.asset) {
    return {
      top: 0,
      left: 0,
      bottom: 0,
      right: 0,
      width: 0,
      height: 0
    };
  }
  const ref = image.asset._ref,
    dimensions = ref?.split('-')?.[2]?.split('x'),
    crop: ImageCrop = {
      top: Math.floor(dimensions[1] * (image?.crop?.top ?? 0)),
      left: Math.floor(dimensions[0] * (image?.crop?.left ?? 0)),
      bottom: Math.floor(dimensions[1] * (image?.crop?.bottom ?? 0)),
      right: Math.floor(dimensions[0] * (image?.crop?.right ?? 0))
    } as ImageCrop;

  crop.width = Math.floor(dimensions[0] - (crop.left + crop.right));
  crop.height = Math.floor(dimensions[1] - (crop.top + crop.bottom));

  return crop;
};
