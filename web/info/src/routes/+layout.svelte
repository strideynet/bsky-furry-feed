<script lang="ts">
  import '../styles/app.scss';

  import { onMount } from 'svelte';

  import { browser } from '$app/environment';
  import { agent, session } from '$lib/atp';

  import PageTransitionWrapper from '$components/layouts/page-transition.svelte';
  import Nav from '$components/nav.svelte';

  import type { LayoutData } from './$types';

  export let data: LayoutData;

  let pageContainer: HTMLDivElement,
    isAtTop = true;

  const checkScroll = (_e?: Event) =>
    (isAtTop = pageContainer?.getBoundingClientRect().top >= 0);

  onMount(() => {
    checkScroll();
  });

  $: hasSession = !!($session !== null && $agent?.hasSession);
  $: ({ pathname } = data.url ?? { pathname: '' });
  $: browser && setTimeout(() => checkScroll(), 100), [pathname];
</script>

<svelte:window on:scroll={checkScroll} on:wheel={checkScroll} />

<svelte:head>
  <title>furryli.st</title>
  <meta name="description" content="The purremier furry feed for Bluesky" />
</svelte:head>

<div class="relative flex flex-col" bind:this={pageContainer}>
  <Nav {hasSession} {isAtTop} />
  <PageTransitionWrapper key={pathname}>
    <div class="w-full px-6 pb-4">
      <slot />
    </div>
  </PageTransitionWrapper>
</div>
