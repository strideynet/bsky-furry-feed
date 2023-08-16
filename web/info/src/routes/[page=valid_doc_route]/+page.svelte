<script lang="ts">
  import StaticPage from '$components/layouts/static-page.svelte';
  import PortableText from '$components/portable-text/portable-text.svelte';

  export let data;

  $: ({ feeds } = data);
  $: content = data.content?.body ?? [];
  $: title = data.content?.title;
</script>

<svelte:head>
  <title>furryli.st{title ? ` | ${title}` : ''}</title>
</svelte:head>

<StaticPage>
  <PortableText {content} {feeds} />
</StaticPage>
