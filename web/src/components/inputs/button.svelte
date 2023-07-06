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
  class="mb-4 w-full rounded-md border-2 border-gray-400/50 bg-gray-50 px-4 py-2 transition-colors duration-75 focus-within:border-gray-500 focus-within:bg-gray-200 focus-within:outline-none hover:border-gray-500 hover:bg-gray-200"
  ><slot /></button
>
