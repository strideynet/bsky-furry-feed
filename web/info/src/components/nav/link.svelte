<script lang="ts">
  import { page } from '$app/stores';

  export let href: string,
    target = '';

  $: isActive = $page?.url.pathname.startsWith(href);
</script>

<a class:isActive {href} {target}><slot /></a>

<style lang="scss">
  a {
    @apply block h-fit whitespace-nowrap text-gray-800 underline-offset-2;

    &:hover,
    &:focus-visible,
    &.isActive {
      @apply text-gray-900 underline;
    }
  }
</style>
