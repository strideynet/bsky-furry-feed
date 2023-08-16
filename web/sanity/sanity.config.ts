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
  schema: {
    types: schemaTypes
  },
  studio: {
    components: {
      logo: Logo
    }
  }
});
