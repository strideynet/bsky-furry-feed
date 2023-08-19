<script lang="ts">
  import { toPlainText } from '@portabletext/svelte';

  import type { BlockComponentProps } from '@portabletext/svelte';

  export let portableText: BlockComponentProps;

  $: ({ value } = portableText);
  $: ({ style } = value as { style: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' });
  $: plainText = toPlainText(value);
</script>

<svelte:element this={style}>
  <a id={value._key} href={`#${value._key}`}>
    {plainText}
  </a>
</svelte:element>

<style lang="scss">
  h1,
  h2,
  h3,
  h4,
  h5 {
    @apply font-bold text-gray-950 dark:text-gray-300;
  }

  h1 {
    @apply mb-6 mt-8 text-4xl;
  }

  h2 {
    @apply mb-5 mt-7 text-3xl;
  }

  h3 {
    @apply mb-4 mt-6 text-2xl;
  }

  h4 {
    @apply mb-3 mt-5 text-xl;
  }

  h5 {
    @apply mb-2 mt-4 text-lg;
  }
</style>
