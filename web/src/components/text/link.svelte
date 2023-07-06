<script lang="ts">
  export let href: string,
    rel = 'noopener noreferrer',
    target = '_blank';
</script>

<a
  {href}
  {rel}
  {target}
  class="text-sky-600 underline-offset-1 focus-within:underline hover:text-sky-700 hover:underline"
  ><slot /></a
>
