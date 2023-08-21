import { PublicService } from '$api/public_service_connectweb';
import { API_URL } from '$lib/constants';

import { createPromiseClient } from '@bufbuild/connect';
import { createConnectTransport } from '@bufbuild/connect-web';

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
