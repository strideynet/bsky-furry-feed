<script lang="ts">
  export let hasSession: boolean,
    isAtTop = false;
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
      {#if hasSession}
        <a href="/dash">Dashboard</a>
        <a href="/auth/logout">Logout</a>
      {:else}
        <a href="/auth/login">Login</a>
      {/if}
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
