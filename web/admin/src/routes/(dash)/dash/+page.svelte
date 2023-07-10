<script lang="ts">
  import { onMount } from 'svelte';

  import { session } from '$lib/atp';
  import { head } from '$stores/head';

  import Heading from '$components/text/heading.svelte';

  onMount(() => head.set({ ...$head, title: 'Dashboard' }));
</script>

<Heading level={1}>Dashboard</Heading>

<p>
  Authenticated as<br />
  Email: <code>{$session?.email}</code><br />
  Identifier: <code>{$session?.handle}</code>
</p>

<style lang="scss">
  code {
    @apply rounded-md bg-gray-200/50 px-1 py-0.5 text-sm;
  }
</style>
