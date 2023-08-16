<script lang="ts">
  import { page } from '$app/stores';

  import StaticPage from '$components/layouts/static-page.svelte';
  import Heading from '$components/text/heading.svelte';

  $: message = $page.error?.message
    ? $page.error.message === 'Not Found'
      ? 'Resource could not be found or does not exist.'
      : $page.error.message
    : 'An unknown error occurred.';
</script>

<StaticPage>
  <Heading level={2}>
    Error {$page.status ?? 500}
  </Heading>

  <p>{message}</p>
</StaticPage>
