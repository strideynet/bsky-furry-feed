import type { DOC_ROUTES_NAMES } from '$lib/constants';

export type { Feed as FeedInfo } from '$api/public_service_pb';

export type BodyBlock = {
  _key: string;
  _type: string;
  children: {
    _key: string;
    _type: string;
    text: string;
    marks: string[];
  }[];
  markDefs: {
    _key: string;
    _type: string;
  }[];
};

export type BaseDocument<T extends string> = {
  _id: T;
  _type: string;
  _createdAt: string;
  _updatedAt: string;
  title: string;
  id: T;
  body: BodyBlock[];
};

type ValidDocs = typeof DOC_ROUTES_NAMES[keyof typeof DOC_ROUTES_NAMES];

export type DocumentRegistry = {
  [key in ValidDocs]: BaseDocument<key>;
} & {
  home: BaseDocument<'home'>;
};

export interface SanityAsset {
  _id: string;
  _type?: string;
  _createdAt?: string;
  _rev?: string;
  _updatedAt?: string;
  url?: string;
  path?: string;
  assetId?: string;
  extension?: string;
}
export interface SanityReference {
  _ref: string;
  _type: string;
}
export interface SanityImageObject extends Pick<SanityAsset, '_id' | '_type'> {
  asset: SanityReference;
  crop?: SanityImageCrop;
  hotspot?: SanityImageHotspot;
}
export interface SanityImageCrop {
  _type?: string;
  left: number;
  bottom: number;
  right: number;
  top: number;
}
export interface SanityImageHotspot {
  _type?: string;
  width: number;
  height: number;
  x: number;
  y: number;
}
