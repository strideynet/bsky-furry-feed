<script lang="ts">
  import '../styles/app.scss';

  import { onMount } from 'svelte';
  import { classList } from 'svelte-body';

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

  onMount(() => checkScroll());

  $: hasSession = !!($session !== null && $agent?.hasSession);
  $: ({ pathname } = data.url ?? { pathname: '' });
  $: browser && setTimeout(() => checkScroll(), 100), [pathname];
</script>

<svelte:window on:scroll={checkScroll} on:wheel={checkScroll} />

<svelte:body use:classList={'light bg-gray-100 dark:text-white dark:bg-gray-800'} />

<svelte:head>
  <title>furryli.st</title>
  <meta name="description" content="The purremier furry feed for Bluesky" />

  <script>
    // Prevent FOUC
    if (
      localStorage.theme === 'dark' ||
      (!('theme' in localStorage) &&
        window.matchMedia('(prefers-color-scheme: dark)').matches)
    ) {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    } else {
      document.documentElement.classList.add('light');
      document.documentElement.classList.remove('dark');
    }
  </script>
</svelte:head>

<div class="relative flex flex-col" bind:this={pageContainer}>
  <Nav {hasSession} {isAtTop} />
  <PageTransitionWrapper key={pathname}>
    <div class="w-full px-6 py-4">
      <slot />
    </div>
  </PageTransitionWrapper>
</div>
