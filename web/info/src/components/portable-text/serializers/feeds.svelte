<script lang="ts">
  import FeaturedList from '$components/feeds/featured-list.svelte';
  import FeedList from '$components/feeds/list.svelte';

  import type { CustomBlockComponentProps } from '@portabletext/svelte';
  import type { FeedInfo } from '$types';

  export let portableText: CustomBlockComponentProps<{ featured: boolean }>;

  $: ({ featured, feeds } = (portableText.global.context?.feeds as {
    featured: FeedInfo[];
    feeds: FeedInfo[];
  }) || {
    featured: null,
    feeds: null
  });
</script>

{#if portableText.value.featured}
  <FeaturedList {feeds} {featured} />
{:else}
  <FeedList {feeds} />
{/if}
