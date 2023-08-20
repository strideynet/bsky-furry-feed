<script lang="ts">
  export let href: string,
    internal = false,
    rel = internal ? undefined : 'noopener noreferrer',
    target = internal ? undefined : '_blank';
</script>

<a
  {href}
  {rel}
  {target}
  class="text-sky-600 underline-offset-1 focus-within:underline hover:text-sky-700 hover:underline"
  ><slot /></a
>
