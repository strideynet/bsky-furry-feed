<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  import { page } from '$app/stores';

  const dispatch = createEventDispatcher();

  export let text: string | undefined,
    href: string,
    target = '';

  $: isActive = $page?.url.pathname.startsWith(href);
</script>

<a
  class:isActive
  {href}
  {target}
  on:click={(e) => dispatch('click', e)}
  on:keydown={(e) => dispatch('keydown', e)}
>
  {#if $$slots.default}
    <slot />
  {:else}
    {text}
  {/if}
</a>

<style lang="scss">
  a {
    @apply block h-fit whitespace-nowrap text-gray-800 dark:text-gray-200 underline-offset-2;

    &:hover,
    &:focus-visible,
    &.isActive {
      @apply text-gray-900 dark:text-gray-100 underline;
    }
  }
</style>
