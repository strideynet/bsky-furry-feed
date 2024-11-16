import NodeAdapter from '@sveltejs/adapter-node';
import CloudflareAdapter from '@sveltejs/adapter-cloudflare';
import { vitePreprocess } from '@sveltejs/kit/vite';
import Preprocess from 'svelte-preprocess';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: [
    vitePreprocess(),
    Preprocess({
      postcss: true
    })
  ],

  kit: {
    alias: {
      $api: '../proto/bff/v1',
      $components: 'src/components',
      $stores: 'src/stores',
      $routes: 'src/routes'
    },
    adapter:
      process.env.SK_ADAPTER === 'cloudflare'
        ? CloudflareAdapter()
        : NodeAdapter({ out: './dist' }),
    files: {
      lib: 'src/lib',
      params: 'src/params',
      routes: 'src/routes'
    }
  }
};

export default config;
