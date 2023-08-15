<script lang="ts">
  import FeaturedFeedsList from '$components/feeds/featured-list.svelte';
  import StaticPage from '$components/layouts/static-page.svelte';
  import Heading from '$components/text/heading.svelte';
  import Link from '$components/text/link.svelte';

  export let data;
</script>

<StaticPage>
  <Heading level={2}>Welcome</Heading>

  <p>
    So. You're a furry, you just joined BlueSky, and you clicked a link somewhere in some
    thread to land you here?
  </p>

  <p>
    Well, you're in the right spot. Let's get you setup and using BlueSky the way you want
    to use it.
  </p>

  <Heading level={3}>Content Filtering</Heading>
  <p>
    BlueSky filters out material it detects as NSFW by default. If you'd like to change
    that, open the "Moderation" menu in your client, click "Content Filtering" and change
    your settings to what you'd like to see. If you're using iOS, you'll need to first
    enable adult content on the web version of BlueSky before you can make changes on iOS.
  </p>

  <p>
    The BlueSky NSFW content detection can be a little inaccurate at times. Manual tagging
    of posts is now present in the client, so if you're posting anything suggestive or
    NSFW, please mark it as such! Adding #nsfw to posts will also serve to hide them from
    the SFW furryli.st feeds.
  </p>

  <Heading level={3}>Our Feeds</Heading>
  <p>
    BlueSky has a concept called "Feeds", which are sort of like Twitter's "For You" and
    "Following" tabs. Except BlueSky's feeds are way more customizable. These feeds can be
    created by anyone and published for other people to use.
  </p>

  <p>
    Here at furryli.st, we build furry-focussed feeds for BlueSky. Our feeds only include
    content from approved "furries" and we offer several more specific feeds that narrow
    down on certain content. These feeds can be a great way to discover other furs on the
    platform and even seek out artists who are open to being commissioned.
  </p>

  <p>
    Below are some of our feeds. To join a feed, just follow the link and click the little
    "pin" at the top of the page (We'd also really appreciate it if you'd hit the like
    button too 😉).
  </p>

  <FeaturedFeedsList feeds={data.feeds} featured={data.featuredFeeds} />

  <Heading level={4}>Joining The Furryli.st Feeds</Heading>

  <p>
    Our feeds are curated to show only furry accounts. To do that we manually verify each
    and every account that gets shown in our feeds. If you'd like to get verified, all you
    need to do is follow <Link href="https://bsky.app/profile/furryli.st"
      >@furryli.st</Link
    > on BlueSky. You'll know you're verified when we follow you back.
  </p>

  <p>
    Once you're verified your posts will start showing up in "New" and "Hot" immediately.
    In order to get your posts included in the tagged feeds, make sure to include the
    appropriate tag in your post. Currently, the available tags are:
  </p>

  <ul>
    <li>#fursuit</li>
    <li>#art / #furryart</li>
    <li>#commsopen</li>
    <li>#nsfw</li>
    <li>#murrsuit</li>
  </ul>

  <p>
    If you still haven't been verified after a day or two, reach out on Discord and we'll
    look into your account for you.
  </p>

  <Heading level={4}>Questions? Problems?</Heading>

  <p>
    We can answer questions and look into problems with furryli.st for you in our <Link
      href="https://discord.gg/7X467r4UXF">Discord</Link
    >. Please try to work out if your question is specific to the feeds or is related to
    BlueSky in general, we don't have the time to answer questions related to BlueSky
    generally.
  </p>

  <Heading level={4}>Supporting Our Feeds</Heading>

  <p>
    Running these feeds for the community isn't free! We'd welcome donations to help with
    the cost of running the feeds on our <Link href="https://ko-fi.com/furrylist"
      >Ko-Fi</Link
    >
  </p>

  <Heading level={3}>Community Resources</Heading>
  <p>There's a bunch of great resources available for furs who are new to BlueSky.</p>

  <p>
    <Link href="https://discord.gg/fursky">FurSky</Link> is an online discord community of
    furs who are on BlueSky - it's a great place to ask BlueSky questions that aren't related
    to the furryli.st feeds!
  </p>

  <p>
    You can find more information in <Link
      href="https://bsky.app/profile/kamu.bsky.social/post/3k3w6w3q4by2b"
      >Kamu's "Furry Newbie Orientation"</Link
    >. Kamu also includes a few more options to get your account set up that aren't
    related to feeds, make sure to go check it out!
  </p>
</StaticPage>

<style lang="scss">
  p {
    @apply my-3;
  }

  ul {
    @apply my-3 ml-4 list-disc;

    li {
      @apply my-0.5 list-item list-inside;
    }
  }
</style>
