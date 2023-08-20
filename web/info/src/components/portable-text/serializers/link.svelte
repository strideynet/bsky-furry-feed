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
  class="text-sky-600 underline-offset-1 transition-[color] duration-75 hover:text-sky-700 hover:underline focus-visible:text-sky-700 focus-visible:underline dark:text-sky-500 dark:hover:text-sky-300 dark:focus-visible:text-sky-300"
  {href}
  target={newtab ? '_blank' : undefined}
  rel={newtab ? 'noopener noreferrer' : undefined}
  tabindex="0"
>
  <slot />
</a>
