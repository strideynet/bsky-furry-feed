<script lang="ts">
  import type { ListItemComponentProps } from '@portabletext/svelte';

  export let portableText: ListItemComponentProps;

  $: ({ value: _value } = portableText);
</script>

<li>
  <slot />
</li>
