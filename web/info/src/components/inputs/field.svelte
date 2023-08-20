<script lang="ts">
  export let type = 'text',
    name = '',
    title = '',
    placeholder: string,
    required = false,
    value = '',
    disabled = false;

  const id = 'input-' + name.replace(/[^a-z0-9]/gi, '');
</script>

<span class="w-full">
  {#if title}
    <label for={id} class="mb-2 ml-1 block self-start text-lg font-bold">{title}</label>
  {/if}
  <input
    {type}
    {name}
    {placeholder}
    {required}
    {value}
    {disabled}
    {id}
    class="my-1 w-full rounded-md bg-gray-50 px-4 py-2 transition-[background-color,box-shadow] duration-75 focus-within:outline-none dark:bg-gray-800/20 dark:focus-within:border-gray-300/50"
  />
</span>

<style lang="scss">
  input {
    @apply from-gray-300/50;
    box-shadow: 0 0 0 2px var(--tw-gradient-from);

    &:focus-within {
      box-shadow: 0 0 0 4px var(--tw-gradient-from);
    }
  }

  :global(.dark) {
    input {
      @apply from-gray-500/50;
    }
  }
</style>
