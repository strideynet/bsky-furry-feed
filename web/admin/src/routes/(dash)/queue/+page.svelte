<script lang="ts">
  import { onMount } from 'svelte';

  import { head } from '$stores/head';

  import QueueItem from '$components/queue/item.svelte';
  import Heading from '$components/text/heading.svelte';

  import type { PageData } from './$types';

  export let data: PageData;

  onMount(() => head.set({ ...$head, title: 'Queue' }));
</script>

<Heading level={1}>Approval Queue</Heading>

{#if !data?.queue?.length}
  <p>There are no pending approvals.</p>
{:else}
  <div class="flex flex-col gap-4">
    {#each data.queue as _item}
      <QueueItem />
    {/each}
  </div>
{/if}
