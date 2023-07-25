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
    class="w-full rounded-md border-2 border-gray-400/50 bg-gray-50 px-4 py-2 transition-colors duration-75 focus-within:border-gray-500 focus-within:outline-none"
  />
</span>
