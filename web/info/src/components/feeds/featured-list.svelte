<script lang="ts">
  import FeedsList from '$components/feeds/list.svelte';
  import ViewMore from '$components/feeds/view-more.svelte';

  import type { FeedInfo } from '$types';

  export let feeds: FeedInfo[], featured: FeedInfo[];
</script>

<FeedsList feeds={featured}>
  {#if featured.length !== feeds.length}
    <ViewMore />
  {/if}
</FeedsList>
