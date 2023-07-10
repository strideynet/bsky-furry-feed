import { writable } from 'svelte/store';

export const head = writable({
  title: '',
  description: ''
});
