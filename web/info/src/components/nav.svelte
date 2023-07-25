<script lang="ts">
  import NavLink from '$components/nav/link.svelte';

  export let hasSession: boolean,
    isAtTop = false;

  $: navLinks = hasSession
    ? [
        { href: '/dash', text: 'Dashboard' },
        { href: '/auth/logout', text: 'Logout' }
      ]
    : [{ href: '/auth/login', text: 'Login' }];
</script>

<div
  class="sticky left-0 top-0 w-screen border-b-4 transition-colors duration-75"
  class:isAtTop
>
  <div
    class="flex h-fit w-full flex-col flex-wrap items-center justify-center gap-6 px-8 py-6 md:flex-row md:justify-start"
  >
    <a class="w-fit text-2xl font-bold md:mb-1" href="/">🐕 furryli.st</a>
    <div class="flex flex-1 flex-row flex-wrap justify-between gap-6 md:flex-nowrap">
      <div class="flex flex-row gap-4">
        <NavLink href="/community-guidelines">Community Guidelines</NavLink>
        <NavLink href="https://discord.gg/7X467r4UXF" target="_blank">Discord</NavLink>
      </div>
      <div class="flex flex-row gap-4">
        {#each navLinks as link}
          <NavLink href={link.href}>{link.text}</NavLink>
        {/each}
      </div>
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
