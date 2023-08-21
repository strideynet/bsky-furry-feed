import { defineConfig } from 'sanity';
import { deskTool } from 'sanity/desk';
import { visionTool } from '@sanity/vision';
import { schemaTypes } from './schemas';
import { structure } from './structure';
import { Logo } from './components/logo';

export default defineConfig({
  name: 'default',
  title: 'bff',
  projectId: '0ildj6pc',
  dataset: 'production',
  plugins: [
    deskTool({
      structure
    }),
    visionTool({
      defaultApiVersion: 'v2022-11-29'
    })
  ],
  document: {
    productionUrl: async (prev, context) => {
      const { getClient, document } = context;

      const client = getClient({
        apiVersion: 'v2022-11-29'
      });

      if (document._type !== 'page') {
        return prev;
      }

      const baseUrl = 'https://furryli.st';

      const params = new URLSearchParams();
      params.set('preview', 'true');
      params.set('token', client.config()?.token || '');

      const id = document._id.replace(/^drafts\./, '');

      switch (id) {
        case 'home':
          return `${baseUrl}/?${params.toString()}`;
        case 'welcome':
          return `${baseUrl}/welcome?${params.toString()}`;
        case 'feeds':
          return `${baseUrl}/feeds?${params.toString()}`;
        case 'communityGuidelines':
          return `${baseUrl}/community-guidelines?${params.toString()}`;
        default:
          return prev;
      }
    }
  },
  schema: {
    types: schemaTypes
  },
  studio: {
    components: {
      logo: Logo
    }
  }
});
