import { redirect } from '@sveltejs/kit';

export function load() {
  throw redirect(302, 'https://discord.com/invite/9aMHDCYnT6');
}
