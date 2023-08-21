<script lang="ts">
  import type { ListComponentProps } from '@portabletext/svelte';

  export let portableText: ListComponentProps;

  $: ({ value: _value } = portableText);
</script>

<ul class="pt-ul my-6 ml-6 list-disc">
  <slot />
</ul>
