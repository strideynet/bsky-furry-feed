<script lang="ts">
  import type { MarkComponentProps } from '@portabletext/svelte';

  export let portableText: MarkComponentProps;

  $: ({ markType } = portableText);
  $: ({ plainTextContent } = portableText);
</script>

{#if markType === 'code'}
  <code class="mx-0.5 rounded bg-gray-400 px-1 py-0.5 font-mono text-[0.95rem]"
    >{plainTextContent}</code
  >
{:else}
  <slot />
{/if}
