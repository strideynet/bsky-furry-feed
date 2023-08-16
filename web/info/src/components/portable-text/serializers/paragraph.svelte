<script lang="ts">
  import type { BlockComponentProps } from '@portabletext/svelte';

  export let portableText: BlockComponentProps;

  $: ({ value } = portableText);
</script>

<p class="my-3 text-base">
  <slot />
</p>
