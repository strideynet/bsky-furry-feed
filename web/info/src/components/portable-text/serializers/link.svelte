<script lang="ts">
  import type { MarkComponentProps } from '@portabletext/svelte';

  export let portableText: MarkComponentProps;

  $: ({ value } = portableText);
  $: ({ plainTextContent } = portableText);
  $: href = value.href as string;
  $: _external = value.external as boolean;
  $: newtab = value.blank as boolean;
</script>

<a
  class="text-sky-600 underline-offset-1 focus-within:text-sky-700 focus-within:underline hover:text-sky-700 hover:underline"
  {href}
  target={newtab ? '_blank' : undefined}
  rel={newtab ? 'noopener noreferrer' : undefined}
  tabindex="0"
>
  <slot />
</a>
