import { VALID_DOC_ROUTES } from '$lib/constants';

import type { ParamMatcher } from '@sveltejs/kit';

export const match: ParamMatcher = (param) => {
  return VALID_DOC_ROUTES.includes(param);
};
