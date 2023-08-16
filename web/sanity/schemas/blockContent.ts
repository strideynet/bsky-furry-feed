import { BsLink45Deg } from 'react-icons/bs';
import { defineType, defineArrayMember } from 'sanity';

/**
 * This is the schema definition for the rich text fields used for
 * for this studio. When you import it in schemas.js it can be
 * reused in other parts of the studio with:
 *  {
 *    name: 'someName',
 *    title: 'Some title',
 *    type: 'bodyContent'
 *  }
 */
export default defineType({
  title: 'Body Content',
  name: 'bodyContent',
  type: 'array',
  of: [
    defineArrayMember({
      title: 'Block',
      type: 'block',
      styles: [
        {title: 'Normal', value: 'normal'},
        {title: 'H1', value: 'h1'},
        {title: 'H2', value: 'h2'},
        {title: 'H3', value: 'h3'},
        {title: 'H4', value: 'h4'},
        {title: 'H5', value: 'h5'},
        {title: 'Quote', value: 'blockquote'}
      ],
      lists: [
        {
          title: 'Bullet',
          value: 'bullet'
        },
        {
          title: 'Numbered',
          value: 'number'
        }
      ],
      marks: {
        decorators: [
          {title: 'Emphasis', value: 'em'},
          {title: 'Strong', value: 'strong'},
          {title: 'Underline', value: 'underline'},
          { title: 'Strikethrough', value: 'strike-through' },
          { title: 'Code', value: 'code' }
        ],
        annotations: [
          {
            title: 'URL',
            name: 'link',
            type: 'object',
            icon: BsLink45Deg,
            fields: [
              {
                title: 'URL',
                name: 'href',
                type: 'string',
                // validation should allow both URIs and relative links (/welcome)
                validation: (Rule) => Rule.required().custom((href: string) => {
                  if (href.startsWith('/')) {
                    return true;
                  }
                  if (href.startsWith('http://') || href.startsWith('https://')) {
                    return true;
                  }
                  return 'Must be a relative link (/welcome) or a URL (https://sanity.io)';
                })
              },
              {
                title: 'Open in new tab',
                name: 'blank',
                type: 'boolean',
                initialValue: false
              },
              {
                title: 'External',
                name: 'external',
                type: 'boolean',
                initialValue: true
              }
            ]
          }
        ]
      }
    }),
    defineArrayMember({
      type: 'image',
      options: {
        hotspot: true
      }
    }),
    defineArrayMember({
      type: 'feeds'
    })
  ]
});
