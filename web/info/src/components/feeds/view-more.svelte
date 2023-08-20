<a
  class="flex w-full flex-1 cursor-pointer flex-col items-center justify-center rounded-md border-4 border-transparent px-4 py-3 outline-2 outline-offset-4 outline-sky-400 transition-[background-color,box-shadow] duration-75 hover:bg-gray-200/50 focus-visible:bg-gray-200/50 dark:hover:bg-gray-800/50 dark:focus-visible:bg-gray-800/50 md:w-auto md:max-w-[320px] xl:max-w-[260px]"
  href="/feeds"
  tabindex="0"
>
  <p class="whitespace-nowrap text-sm text-sky-600 dark:text-sky-500">View more...</p>
</a>

<style lang="scss">
  a {
    @apply from-gray-300/50;
    box-shadow: 0 0 0 2px var(--tw-gradient-from);

    &:hover,
    &:focus-visible {
      box-shadow: 0 0 0 4px var(--tw-gradient-from);

      p {
        @apply text-sky-700 underline;
      }
    }
  }

  :global(.dark) {
    a {
      @apply from-gray-500/50;
      box-shadow: 0 0 0 2px var(--tw-gradient-from);

      &:hover,
      &:focus-visible {
        box-shadow: 0 0 0 4px var(--tw-gradient-from);

        p {
          @apply text-sky-300;
        }
      }
    }
  }
</style>
