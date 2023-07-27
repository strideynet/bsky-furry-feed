<script lang="ts">
  export let level: 1 | 2 | 3 | 4 | 5 = 1,
    id: string | undefined = undefined,
    classNames = '';
</script>

<svelte:element this={`h${level}`} class={`heading-${level} ${classNames}`} {id}
  ><slot /></svelte:element
>

<style lang="scss">
  .heading {
    &-1 {
      @apply mb-5 mt-6 text-4xl font-bold;
    }
    &-2 {
      @apply mb-4 mt-5 text-3xl font-bold;
    }
    &-3 {
      @apply mb-3 mt-4 text-2xl font-bold;
    }
    &-4 {
      @apply mb-2 mt-3 text-xl font-bold;
    }
    &-5 {
      @apply mb-1 mt-2 text-lg font-bold;
    }
  }
</style>
