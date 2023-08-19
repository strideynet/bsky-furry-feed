<script lang="ts">
  import Heading from '$components/text/heading.svelte';

  import type { FeedInfo } from '$types';

  export let feed: FeedInfo;

  $: tags = feed.tags?.length
    ? feed.tags.map((tag) => '#' + tag.toLowerCase()).join(', ')
    : undefined;
</script>

<a
  class="flex w-full flex-auto cursor-pointer flex-col gap-y-1 rounded-md border-4 border-transparent px-4 pb-3 pt-1 outline-2 outline-offset-4 outline-sky-400 transition-[background-color,box-shadow] duration-75 hover:bg-gray-100 focus-visible:bg-gray-100 dark:hover:bg-gray-700 dark:focus-visible:bg-gray-700 md:w-auto md:max-w-[320px] xl:max-w-[260px]"
  href={feed.link}
  target="_blank"
  tabindex="0"
  id={`feed-${feed.id}`}
>
  <Heading level={5}>{feed.displayName}</Heading>
  {#if tags}
    <p class="-my-1 mb-0.5 text-sm text-gray-500">
      {tags}
    </p>
  {/if}
  <p class="text-sm">{feed.description}</p>
</a>

<style lang="scss">
  a {
    box-shadow: 0 0 0 2px rgb(209 213 219 / 0.5);

    &:hover,
    &:focus-visible {
      box-shadow: 0 0 0 4px rgb(209 213 219 / 0.5);
    }
  }
</style>
