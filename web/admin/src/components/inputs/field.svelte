<script lang="ts">
  export let type = 'text',
    name = '',
    placeholder = '',
    required = false,
    value = '',
    disabled = false;
</script>

<input
  {type}
  {name}
  {placeholder}
  {required}
  {value}
  {disabled}
  class="mb-4 w-full rounded-md border-2 border-gray-400/50 bg-gray-50 px-4 py-2 transition-colors duration-75 focus-within:border-gray-500 focus-within:outline-none"
/>
