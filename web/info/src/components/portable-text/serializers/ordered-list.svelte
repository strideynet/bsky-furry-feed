<script lang="ts">
  import type { BlockComponentProps } from '@portabletext/svelte';

  export let portableText: BlockComponentProps;

  const getLevelClass = (level?: number) => {
    switch (level) {
      case 2:
        return 'list-[lower-alpha]';
      case 3:
        return 'list-[lower-roman]';
      default:
        return 'list-decimal';
    }
  };

  $: ({ value } = portableText);
  $: ({ level } = value);
</script>

<ol class="pt-ol my-6 ml-6 {getLevelClass(level)}">
  <slot />
</ol>
