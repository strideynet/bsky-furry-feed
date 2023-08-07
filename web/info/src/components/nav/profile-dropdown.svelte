<script lang="ts">
  import { profile } from '$lib/atp';

  import {
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
    Transition
  } from '@rgossiaux/svelte-headlessui';

  export let hasSession: boolean;

  $: profileAvatarSrc =
    $profile && $profile.avatar ? $profile.avatar : '/placeholder-avatar.png';
</script>

<Menu class="relative">
  <MenuButton aria-label="Toggle user menu">
    <img
      src={profileAvatarSrc}
      alt="Avatar"
      class="h-8 w-8 rounded-full border border-gray-500"
    />
  </MenuButton>
  <Transition
    enter="transition duration-100 ease-out"
    enterFrom="transform scale-95 opacity-0"
    enterTo="transform scale-100 opacity-100"
    leave="transition duration-75 ease-out"
    leaveFrom="transform scale-100 opacity-100"
    leaveTo="transform scale-95 opacity-0"
  >
    <MenuItems
      class="absolute right-0 mt-2 flex min-w-[150px] origin-top-right flex-col rounded-md bg-gray-50 px-2 py-1 shadow-[0_0_0_3px_--var(tw-shadow-color)] shadow-gray-300/50 outline-none"
    >
      {#if hasSession}
        <MenuItem disabled>
          <span class="my-1 block w-full cursor-default p-2 text-left">
            <span class="text-gray-600">@{$profile?.handle}</span>
          </span>
        </MenuItem>
        <span class="block h-[1px] bg-gray-300/50" />
        <MenuItem let:active class="mb-0.5 mt-1.5">
          <a class:active href="/dash"> Dashboard </a>
        </MenuItem>
        <MenuItem let:active class="mb-1.5 mt-0.5">
          <a class:active href="/auth/logout"> Logout </a>
        </MenuItem>
      {:else}
        <MenuItem let:active class="my-1.5">
          <a class:active href="/auth/login"> Login </a>
        </MenuItem>
      {/if}
    </MenuItems>
  </Transition>
</Menu>

<style lang="scss">
  a {
    @apply block w-full cursor-pointer rounded-md px-3 py-2 text-left transition-[background-color] duration-75;

    &.active {
      @apply bg-gray-300/50;
    }
  }
</style>
