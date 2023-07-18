<script lang="ts">
  export let level: 1 | 2 | 3 | 4 | 5 = 1;
</script>

{#if level === 1}
  <h1 class="mb-5 mt-6 text-4xl font-bold"><slot /></h1>
{:else if level === 2}
  <h2 class="mb-4 mt-5 text-3xl font-bold"><slot /></h2>
{:else if level === 3}
  <h3 class="mb-3 mt-4 text-2xl font-bold"><slot /></h3>
{:else if level === 4}
  <h4 class="mb-2 mt-3 text-xl font-bold"><slot /></h4>
{:else if level === 5}
  <h5 class="mb-1 mt-2 text-lg font-bold"><slot /></h5>
{/if}
