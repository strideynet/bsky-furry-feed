<script setup lang="ts">
defineProps<{ noPadding?: boolean; variant?: "error" }>();
</script>

<template>
  <div
    class="border border-gray-300 dark:border-gray-700 rounded-lg"
    :class="{
      'px-4 py-3': !noPadding,
      'bg-red-600 bg-opacity-50': variant === 'error',
      'flex gap-2 items-center': variant,
    }"
  >
    <icon-cross-octagon v-if="variant === 'error'" class="flex-shrink-0" />
    <slot />
  </div>
</template>
