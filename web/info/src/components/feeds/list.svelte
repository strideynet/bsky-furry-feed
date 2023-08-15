<script lang="ts">
  import FeedCard from '$components/feeds/card.svelte';

  import type { FeedInfo } from '$types';

  export let feeds: FeedInfo[] = [];
</script>

<div class="my-5 flex flex-col flex-wrap gap-4 md:flex-row">
  {#each feeds as feed}
    <FeedCard {feed} />
  {/each}
  <slot />
</div>
