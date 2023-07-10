import { writable } from 'svelte/store';

import { APP_THEMES } from '$lib/constants';

export default writable<typeof APP_THEMES[keyof typeof APP_THEMES]>(APP_THEMES.light);
