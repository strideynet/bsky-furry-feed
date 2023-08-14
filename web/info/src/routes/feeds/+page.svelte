<script lang="ts">
  import FeedCard from '$components/feeds/card.svelte';
  import StaticPage from '$components/layouts/static-page.svelte';
  import Heading from '$components/text/heading.svelte';

  export let data;
</script>

<StaticPage>
  <Heading level={2}>Our Feeds</Heading>

  <p>
    These are all the feeds we currently operate. Click any of the following links to open
    the feed in your client. After opening the feed, you'll be able to pin it in your
    client. We'd love it if you also gave the feed a like!
  </p>

  <div class="my-5 flex flex-col flex-wrap gap-4 md:flex-row">
    {#each data.feeds as feed}
      <FeedCard {feed} />
    {/each}
  </div>
</StaticPage>

<style lang="scss">
  p {
    @apply mb-3 mt-2;
  }
</style>
