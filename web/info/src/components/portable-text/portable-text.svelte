<script lang="ts">
  import CodeSerializer from '$components/portable-text/serializers/code.svelte';
  import FeedsSerializer from '$components/portable-text/serializers/feeds.svelte';
  import HeadingSerializer from '$components/portable-text/serializers/heading.svelte';
  import ImageSerializer from '$components/portable-text/serializers/image.svelte';
  import LinkSerializer from '$components/portable-text/serializers/link.svelte';
  import OlSerializer from '$components/portable-text/serializers/ordered-list.svelte';
  import OlItemSerializer from '$components/portable-text/serializers/ordered-list-item.svelte';
  import ParagraphSerializer from '$components/portable-text/serializers/paragraph.svelte';
  import UlSerializer from '$components/portable-text/serializers/unordered-list.svelte';
  import UlItemSerializer from '$components/portable-text/serializers/unordered-list-item.svelte';

  import { PortableText } from '@portabletext/svelte';

  import type { InputValue } from '@portabletext/svelte/ptTypes';
  import type { FeedInfo } from '$types';

  export let content: InputValue,
    feeds: { featured: FeedInfo[] | null; feeds: FeedInfo[] | null };
</script>

<PortableText
  value={content}
  components={{
    types: {
      feeds: FeedsSerializer,
      image: ImageSerializer
    },
    marks: {
      link: LinkSerializer,
      code: CodeSerializer
    },
    block: {
      normal: ParagraphSerializer,
      h1: HeadingSerializer,
      h2: HeadingSerializer,
      h3: HeadingSerializer,
      h4: HeadingSerializer,
      h5: HeadingSerializer
    },
    list: {
      bullet: UlSerializer,
      number: OlSerializer
    },
    listItem: {
      bullet: UlItemSerializer,
      number: OlItemSerializer,
      normal: UlItemSerializer
    }
  }}
  context={{
    feeds
  }}
/>
