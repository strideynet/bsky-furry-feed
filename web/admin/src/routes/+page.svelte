<script lang="ts">
  import Heading from '$components/text/heading.svelte';
  import Link from '$components/text/link.svelte';
</script>

<div>
  <Heading level={2}>The purremier furry feed for Bluesky</Heading>

  <p>Ever wanted to scroll a feed of content from furries just like you?</p>

  <p>
    Furryli.st wants to be the best custom feed for furries who use Bluesky. We'll be
    doing our best to show you the freshest furry content.
  </p>

  <Heading level={3}>How do I get started?</Heading>
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
  </ul>

  <Heading level={3}>How does it work?</Heading>

  <p>
    Our moderation team, with the help of a few tools, manually approve new furries to be
    added to the list. Once added, we start tracking the posts, reposts, likes and follows
    of these furries. Using this data we can then compile a feed of furry content!
  </p>

  <p>
    If you want to make sure you are opted-in, or opted-out, please contact us and we'll
    make sure everything is in order.
  </p>

  <Heading level={3}>Moderation policy</Heading>

  <p>
    It's important to us that the furry feed is a welcoming environment all can enjoy.
  </p>

  <p>
    Right now, we haven't formalised our moderation policy, but we intend to do so soon.
    We won't be tolerating hateful content.
  </p>

  <p>To moderate the feed, we'll be making use of a few different actions:</p>

  <ul>
    <li>Hiding a post from our feeds</li>
    <li>
      Hiding all posts by a user from our feeds, either temporarily (a suspension) or
      permanently (a ban)
    </li>
  </ul>

  <Heading level={3}>FAQs</Heading>

  <Heading level={5}>Can I read the code??</Heading>

  <p>Yup - everything is open-source and can be found on GitHub</p>

  <Heading level={5}>I have more questions!</Heading>

  <p>
    First of all, that's not a question! But, yes, you can find our support channel in the <Link
      href="https://discord.gg/5UNyBtnwKy">Bluesky Furry Discord</Link
    >
  </p>
</div>

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
