<script setup lang="ts">
defineEmits<{
  (event: "close"): void;
}>();
</script>

<template>
  <div
    class="fixed flex top-0 left-0 right-0 bottom-0 z-10 justify-center items-center"
  >
    <slot />
    <div
      class="bg-black bg-opacity-50 absolute top-0 left-0 right-0 bottom-0"
      @click="$emit('close')"
    />
  </div>
</template>
