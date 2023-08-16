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
