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
  class="block h-fit whitespace-nowrap text-gray-800 underline-offset-2 transition-[color] duration-75 dark:text-gray-300"
  {href}
  {target}
  on:click={(e) => dispatch('click', e)}
  on:keydown={(e) => dispatch('keydown', e)}
  data-sveltekit-preload-data
  data-sveltekit-preload-code
>
  {#if $$slots.default}
    <slot />
  {:else}
    {text}
  {/if}
</a>

<style lang="scss">
  a {
    &:hover,
    &:focus-visible,
    &.isActive {
      @apply text-gray-900 underline;
    }
  }

  :global(.dark) {
    a {
      &:hover,
      &:focus-visible,
      &.isActive {
        @apply text-gray-100 underline;
      }
    }
  }
</style>
