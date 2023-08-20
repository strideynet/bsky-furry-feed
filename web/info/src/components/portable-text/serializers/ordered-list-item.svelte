<script lang="ts">
  import type { BlockComponentProps } from '@portabletext/svelte';

  export let portableText: BlockComponentProps;

  $: ({ value: _value } = portableText);
</script>

<li>
  <slot />
</li>
