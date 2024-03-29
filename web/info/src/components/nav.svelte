<script lang="ts">
  import { slide } from 'svelte/transition';

  import { NAV_OPTIONS } from '$lib/constants';

  import NavLink from '$components/nav/link.svelte';
  import MenuButton from '$components/nav/menu-button.svelte';
  import NavProfileDropdown from '$components/nav/profile-dropdown.svelte';

  import ThemeButton from './nav/theme-button.svelte';

  export let hasSession: boolean,
    isAtTop = false;

  let navExpanded = false;
</script>

<div
  class="sticky left-0 top-0 w-screen border-b-4 bg-gray-100 transition-colors duration-75 dark:bg-gray-900"
  class:isAtTop
>
  <!-- Mobile nav -->
  <div class="flex w-full flex-col justify-center px-6 py-5 md:hidden">
    <div class="flex flex-row justify-between">
      <a
        class="block w-fit text-2xl font-bold md:mb-1"
        href="/"
        on:click={() => (navExpanded = false)}
        on:keydown={(e) => e.key === 'Enter' && (navExpanded = false)}>🐕 furryli.st</a
      >
      <button
        class="-m-3 block p-3 md:hidden"
        on:click={() => (navExpanded = !navExpanded)}
        on:keydown={(e) => e.key === 'Enter' && (navExpanded = !navExpanded)}
        aria-label="Toggle navigation"
        aria-expanded={navExpanded}
      >
        <MenuButton icon={navExpanded ? 'close' : 'menu'} />
      </button>
    </div>
    {#if navExpanded}
      <div
        class="mt-4 flex w-full flex-col items-end gap-4"
        in:slide={{ duration: 200 }}
        out:slide={{ duration: 200 }}
      >
        <div class="flex flex-row items-center gap-4">
          <ThemeButton />
          <NavProfileDropdown {hasSession} />
        </div>
        {#each NAV_OPTIONS as link}
          <NavLink
            {...link}
            on:click={() => (navExpanded = false)}
            on:keydown={(e) => e.detail.key === 'Enter' && (navExpanded = false)}
          />
        {/each}
      </div>
    {/if}
  </div>

  <!-- Desktop nav -->
  <div
    class="hidden h-fit w-full flex-col flex-wrap items-center justify-center gap-6 px-8 py-6 md:flex md:flex-row md:justify-start"
  >
    <a class="block w-fit text-2xl font-bold md:mb-1" href="/">🐕 furryli.st</a>
    <div class="flex flex-1 flex-row items-center justify-between gap-6">
      <div class="flex flex-row gap-4">
        {#each NAV_OPTIONS as link}
          <NavLink {...link} />
        {/each}
      </div>
      <div class="flex flex-row items-center gap-5">
        <ThemeButton />
        <NavProfileDropdown {hasSession} />
      </div>
    </div>
  </div>
</div>

<style lang="scss">
  div.sticky {
    overscroll-behavior: none;

    &:not(.isAtTop) {
      @apply border-gray-300/50;
    }
    &.isAtTop {
      @apply border-transparent;
    }
  }

  :global(.dark) {
    div.sticky {
      &:not(.isAtTop) {
        @apply border-gray-500/50;
      }
    }
  }
</style>
