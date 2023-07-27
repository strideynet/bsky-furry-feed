<script setup lang="ts">
defineProps<{ noPadding?: boolean }>();
</script>

<template>
  <div
    class="border border-gray-300 dark:border-gray-700 rounded-lg"
    :class="{ 'px-4 py-3': !noPadding }"
  >
    <slot />
  </div>
</template>
