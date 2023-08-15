<script lang="ts">
  import FeedCard from '$components/feeds/card.svelte';

  import type { FeedInfo } from '$types';

  export let feeds: FeedInfo[] | null = null;
</script>

<div class="my-5 flex flex-col flex-wrap gap-4 md:flex-row">
  {#if feeds}
    {#each feeds as feed}
      <FeedCard {feed} />
    {/each}
  {:else}
    <span class="w-full rounded-md bg-gray-200 px-2 py-1 text-center text-sm"
      >Error fetching feeds</span
    >
  {/if}
  <slot />
</div>
