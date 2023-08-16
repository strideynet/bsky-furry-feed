import {
  BsFillFileRichtextFill,
  BsFillHouseDoorFill,
  BsFillBookmarksFill,
  BsFillShieldLockFill
} from 'react-icons/bs';
import { BiSolidHand } from 'react-icons/bi';
import type { StructureBuilder, StructureResolverContext } from 'sanity/desk';

export const structure = (
  S: StructureBuilder,
  _ctx: StructureResolverContext
) =>
  S.list()
    .title('Content')
    .items([
      S.listItem()
        .title('Home')
        .icon(BsFillHouseDoorFill)
        .child(
          S.document()
            .schemaType('page')
            .documentId('home')
            .title('Home')
        ),
      S.listItem()
        .title('Welcome')
        .icon(BiSolidHand)
        .child(
          S.document()
            .schemaType('page')
            .documentId('welcome')
            .title('Welcome')
        ),
      S.listItem()
        .title('Feeds')
        .icon(BsFillBookmarksFill)
        .child(
          S.document()
            .schemaType('page')
            .documentId('feeds')
            .title('Feeds')
        ),
      S.listItem()
        .title('Community Guidelines')
        .icon(BsFillShieldLockFill)
        .child(
          S.document()
            .schemaType('page')
            .documentId('communityGuidelines')
            .title('Community Guidelines')
        ),
      S.divider(),
      S.listItem()
          .title('All Pages')
          .icon(BsFillFileRichtextFill)
          .child(
            S.documentTypeList('page')
              .title('Pages')
              .filter('_type == "page"')
          )
    ])