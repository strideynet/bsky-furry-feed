import { API_URL } from '$lib/constants';

import { createPromiseClient } from '@bufbuild/connect';
import { createConnectTransport } from '@bufbuild/connect-web';
import { MethodKind, proto3 } from '@bufbuild/protobuf';

const ListFeedsRequest = proto3.makeMessageType('bff.v1.ListFeedsRequest', []);

const ListFeedsResponse = proto3.makeMessageType('bff.v1.ListFeedsResponse', () => [
  { no: 1, name: 'feeds', kind: 'message', T: Feed, repeated: true }
]);

const Feed = proto3.makeMessageType('bff.v1.Feed', () => [
  { no: 1, name: 'id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  { no: 2, name: 'link', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  { no: 3, name: 'display_name', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  { no: 4, name: 'description', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  { no: 5, name: 'priority', kind: 'scalar', T: 5 /* ScalarType.INT32 */ }
]);

const PublicService = {
  typeName: 'bff.v1.PublicService',
  methods: {
    listFeeds: {
      name: 'ListFeeds',
      I: ListFeedsRequest,
      O: ListFeedsResponse,
      kind: MethodKind.Unary
    }
  }
};

type RouteFetch = typeof fetch;

const createTransport = (routeFetch: RouteFetch) =>
  createConnectTransport({
    baseUrl: API_URL,
    fetch: (input, data: RequestInit = {}) => routeFetch(input, { ...data })
  });

const createClient = (routeFetch: RouteFetch) =>
  createPromiseClient(PublicService, createTransport(routeFetch));

let client = null as ReturnType<typeof createClient> | null;

export const getClient = (routeFetch: RouteFetch) => {
  client ??= createClient(routeFetch);
  return client;
};
