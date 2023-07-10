<script lang="ts">
  import NavItem from '$components/nav/item.svelte';

  export let hasSession: boolean,
    isAtTop = false;

  $: navItems = hasSession
    ? [
        { text: 'Dashboard', href: '/dash' },
        { text: 'Queue', href: '/queue' },
        { text: 'Logout', href: '/auth/logout', danger: true }
      ]
    : [{ text: 'Login', href: '/auth/login' }];
</script>

<div
  class="sticky left-0 top-0 w-screen border-b-4 transition-colors duration-75"
  class:isAtTop
>
  <div class="flex h-fit w-full flex-row items-center justify-between px-8 py-6">
    <div class="flex flex-row items-center justify-start gap-4">
      <a class="text-2xl font-bold" href="/">🐕 furryli.st</a>
    </div>
    <div class="flex flex-row gap-4">
      {#each navItems as item}
        <NavItem {...item} />
      {/each}
    </div>
  </div>
</div>

<style lang="scss">
  div.sticky {
    overscroll-behavior: none;

    &:not(.isAtTop) {
      @apply border-gray-300/50 bg-gray-100;
    }
    &.isAtTop {
      @apply border-transparent bg-gray-100;
    }
  }
</style>
