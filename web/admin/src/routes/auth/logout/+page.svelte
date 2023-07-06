<script lang="ts">
  import { onMount } from 'svelte';

  import { goto } from '$app/navigation';
  import { session } from '$lib/atp';

  onMount(async () => {
    session.set(null);
    await new Promise((resolve) => setTimeout(resolve, 500));
    goto('/');
  });
</script>

<p>Logging out...</p>
