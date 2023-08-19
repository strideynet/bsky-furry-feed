<a
  class="flex w-full flex-1 cursor-pointer flex-col items-center justify-center rounded-md border-4 border-transparent px-4 py-3 outline-2 outline-offset-4 outline-sky-400 transition-[background-color,box-shadow] duration-75 hover:bg-gray-100 focus-visible:bg-gray-100 dark:hover:bg-gray-700 dark:focus-visible:bg-gray-700 md:w-auto md:max-w-[320px] xl:max-w-[260px]"
  href="/feeds"
  tabindex="0"
>
  <p class="whitespace-nowrap text-sm text-sky-600">View more...</p>
</a>

<style lang="scss">
  a {
    box-shadow: 0 0 0 2px rgb(209 213 219 / 0.5);

    &:hover,
    &:focus-visible {
      box-shadow: 0 0 0 4px rgb(209 213 219 / 0.5);

      p {
        @apply text-sky-700 underline;
      }
    }
  }
</style>
