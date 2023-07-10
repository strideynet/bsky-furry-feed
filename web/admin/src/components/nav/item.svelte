<script lang="ts">
  import { page } from '$app/stores';

  export let text: string,
    href: string,
    danger = false;

  $: isCurrent = $page?.url.pathname.startsWith(href);
</script>

<a {href} class:isCurrent class:danger>{text}</a>

<style lang="scss">
  a {
    @apply transition-colors duration-75;

    &:hover,
    &:focus-within {
      @apply underline underline-offset-2;
    }

    &.isCurrent {
      @apply font-bold decoration-2;
    }

    &.danger {
      &:hover,
      &:focus-within {
        @apply text-red-700 decoration-red-700;
      }
    }
  }
</style>
