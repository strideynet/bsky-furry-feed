<script lang="ts">
  import { cubicIn, cubicOut } from 'svelte/easing';
  import { fade } from 'svelte/transition';

  import { BASE_TRANSITION_DURATION } from '$lib/constants';

  export let key: string;
</script>

{#key key}
  <div
    in:fade={{
      duration: BASE_TRANSITION_DURATION * 0.8,
      delay: BASE_TRANSITION_DURATION * 0.8,
      easing: cubicIn
    }}
    out:fade={{ duration: BASE_TRANSITION_DURATION * 0.8, easing: cubicOut }}
  >
    <slot />
  </div>
{/key}
