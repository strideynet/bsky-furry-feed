import { defineType } from 'sanity';

export default defineType({
  name: 'feeds',
  title: 'Feeds',
  type: 'object',
  fields: [
    {
      name: 'featured',
      title: 'Featured',
      type: 'boolean',
      description: 'Only show featured feeds',
      initialValue: true
    }
  ],
  preview: {
    select: {
      onlyFeatured: 'featured'
    },
    prepare: ({ onlyFeatured }) => {
      return {
        title: `Feeds: ${onlyFeatured ? 'Only Featured' : 'All'}`
      }
    }
  }
});
