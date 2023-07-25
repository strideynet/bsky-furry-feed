<script lang="ts">
  import StaticPage from '$components/layouts/static-page.svelte';
  import Heading from '$components/text/heading.svelte';
  import Link from '$components/text/link.svelte';
</script>

<StaticPage>
  <Heading level={2}>The purremier furry feed for BlueSky</Heading>

  <p>
    Want to find your fellow furs and view the hottest furry posts on BlueSky? <strong
      >You're in the right place.</strong
    >
  </p>

  <p>
    We operate a set of feeds which serve up furry content straight to your BlueSky
    client. Read on to find out more!
  </p>

  <Heading level={3}>Getting started</Heading>
  <p>Click any of the following links to open the feed in your client:</p>

  <ul>
    <li>
      <Link
        href="https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-new"
        >New</Link
      >
    </li>
    <li>
      <Link
        href="https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-hot"
        >Hot</Link
      >
    </li>
    <li>
      <Link
        href="https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-fursuit"
        >Fursuits</Link
      >
    </li>
    <li>
      <Link
        href="https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-art"
        >Art</Link
      >
    </li>
    <li>
      <Link
        href="https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-nsfw"
        >NSFW</Link
      >
    </li>
    <li>
      <Link
        href="https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/furry-comms"
        >Commissions Open</Link
      >
    </li>
  </ul>

  <p>
    After opening the feed, you'll be able to pin it in your client. We'd love it if you
    also gave the feed a like!
  </p>

  <Heading level={4}>Joining the feed</Heading>

  <p>
    Only approved furries will show up in our feeds. To request approval, follow the
    @furryli.st account. Once you are approved, that account will follow you back.
  </p>

  <p>
    Certain feeds will only show posts including special hashtags. Include these in your
    post to make sure it shows up in the right place. These are:
  </p>
  <ul>
    <li>
      NSFW <ul><li>#nsfw</li></ul>
    </li>
    <li>
      Art <ul><li>#art</li></ul>
    </li>
    <li>
      Fursuits <ul><li>#fursuit</li></ul>
    </li>
    <li>
      Commissions Open <ul><li>#commsopen</li></ul>
    </li>
  </ul>

  <Heading level={3}>Moderation policy</Heading>

  <p>
    Read our <Link href="/community-guidelines">Community Guidelines.</Link>
  </p>

  <Heading level={3}>FAQs</Heading>

  <Heading level={5}>Can I read the code??</Heading>

  <p>
    Yup - everything is open-source and can be found on <Link
      href="https://github.com/strideynet/bsky-furry-feed">GitHub</Link
    >
  </p>

  <Heading level={5}>What else should I check out?</Heading>

  <p>
    BlueSky has a thriving furry community, there's a <Link
      href="https://discord.gg/fursky">"Fursky" Discord</Link
    > that's worth joining to learn more about other resources in the Furry community.
  </p>

  <Heading level={5}>I have more questions!</Heading>

  <p>
    First of all, that's not a question! But, yes, you can find our support channel in <Link
      href="https://discord.gg/7X467r4UXF">our Discord</Link
    >
  </p>
</StaticPage>

<style lang="scss">
  p {
    @apply mb-3 mt-2;
  }

  ul {
    @apply my-3 ml-4 list-disc;

    li {
      @apply my-0.5 list-item list-inside;
    }
  }
</style>
