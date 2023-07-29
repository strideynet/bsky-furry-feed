<script lang="ts">
  import StaticPage from '$components/layouts/static-page.svelte';
  import Heading from '$components/text/heading.svelte';
</script>

<StaticPage>
  <Heading level={2}>Community Guidelines</Heading>

  <p>
    Furryli.st is committed to creating a safe and enjoyable environment for furries. The
    below community guidelines outline how we will decide when to take action against
    content on the feeds.
  </p>

  <p>
    <strong>
      As a general rule, we will be intolerant to intolerance and place a priority on
      protecting marginalised groups.
    </strong>
  </p>

  <Heading level={3}>Rules</Heading>

  <ol>
    <li>
      Do not engage in hate speech:
      <ol>
        <li>Racism</li>
        <li>Homophobia</li>
        <li>Ableism</li>
        <li>Sexism</li>
        <li>Transphobia</li>
        <li>Other forms of prejudice not explicitly listed</li>
      </ol>
    </li>
    <li>
      Do not encourage or engage in harassment of other users. This includes:
      <ol>
        <li>
          Doxxing (posting identifying information about an individual without their
          consent)
        </li>
        <li>Encouraging "dogpiles"</li>
      </ol>
    </li>
    <li>
      Do not sexualise or interact sexually with those under the age of 18. This includes
      posting fictional content portraying characters under the age of 18
    </li>
    <li>Do not endorse or encourage acts of bestiality</li>
    <li>
      Do not attempt to manipulate the feed using automated tooling. For example, creating
      multiple accounts and using an automation to like all your posts using these
      accounts
    </li>
  </ol>

  <Heading level={3}>Enforcement</Heading>

  <p>How these guidelines will be enforced will depend on:</p>
  <ul>
    <li>Which guideline has been broken</li>
    <li>How severely it has been broken</li>
    <li>
      An accounts history of breaking guidelines. Many small infringements may eventually
      be treated as a greater infringement
    </li>
    <li>
      Wider context available about that account, including other posts on BlueSky and
      their posts on other social media platforms
    </li>
  </ul>

  <p>We will make use of the following enforcement actions:</p>
  <ul>
    <li>Removing an individual post from showing on the feeds</li>
    <li>Marking an individual post as NSFW</li>
    <li>Marking all past and future posts by an account as NSFW</li>
    <li>
      Banning an account entirely from the feeds. This means their posts will no longer
      show on the feeds
    </li>
    <li>
      Adding a user to the furryli.st maintained mute/block list. This is reserved for
      accounts we believe pose a danger to the furry community.
    </li>
  </ul>

  <p>
    Where a user has multiple accounts, actions such as bans will take place across all of
    their accounts where appropriate.
  </p>

  <Heading level={3}>Transparency & Complaints</Heading>

  <p>
    In order to allow our moderators to operate effectively without fear of retribution,
    we will not reveal the identity of which moderator took a specific action. Where there
    are concerns with an action that has been taken, or not taken, this should be raised
    with the Director of Furryli.st.
  </p>
</StaticPage>

<style lang="scss">
  p {
    @apply my-3;
  }

  ol {
    @apply my-3 ml-4 list-disc;

    li {
      @apply my-0.5 list-item list-inside;
    }
  }

  ul {
    @apply my-3 ml-4 list-disc;

    li {
      @apply my-0.5 list-item list-inside;
    }
  }
</style>
