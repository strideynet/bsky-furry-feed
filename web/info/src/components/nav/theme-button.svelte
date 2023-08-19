<script lang="ts">
  import { get } from 'svelte/store';

  import { APP_THEMES } from '$lib/constants';
  import { theme } from '$stores/theme';

  const toggle = () => {
    get(theme) === APP_THEMES[0] ? theme.set(APP_THEMES[1]) : theme.set(APP_THEMES[0]);
  };
</script>

<button
  class="-m-3 p-3"
  on:click={toggle}
  on:keyup={(e) => {
    if (e.key === 'Enter') {
      toggle();
    }
  }}
  aria-label="Toggle theme"
>
  toggle theme
</button>
