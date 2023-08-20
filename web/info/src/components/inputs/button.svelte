<script lang="ts">
  export let type: 'button' | 'submit' | 'reset' = 'button',
    name = '',
    value = '',
    disabled = false,
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onClick = () => {};
</script>

<button
  {type}
  {name}
  {value}
  {disabled}
  on:click={onClick}
  on:keyup={(e) => {
    if (e.key === 'Enter') {
      onClick();
    }
  }}
  class="mb-4 mt-1 w-full rounded-md bg-gray-50 px-4 py-2 transition-[background-color,box-shadow] duration-75 hover:bg-gray-200/50 focus-visible:bg-gray-200/50 focus-visible:outline-none dark:bg-gray-800/20 dark:hover:bg-gray-800/50 dark:focus-visible:bg-gray-800/50"
  ><slot /></button
>

<style lang="scss">
  button {
    @apply from-gray-300/50;
    box-shadow: 0 0 0 2px var(--tw-gradient-from);

    &:hover,
    &:focus-visible {
      box-shadow: 0 0 0 4px var(--tw-gradient-from);
    }
  }

  :global(.dark) {
    button {
      @apply from-gray-500/50;
    }
  }
</style>
