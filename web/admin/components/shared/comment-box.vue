<script setup lang="ts">
const $emit = defineEmits<{ (event: "comment", comment: string): void }>();

const text = ref("");

function comment() {
  $emit("comment", text.value);
  text.value = "";
}
</script>

<template>
  <div class="flex gap-2 mt-5 mb-24 text-sm">
    <input
      v-model="text"
      class="flex-1 py-1 px-2 rounded-lg border border-gray-300 text-black"
      placeholder="Type your comment..."
      type="text"
    />
    <button
      :disabled="text.trim().length === 0"
      class="px-3 py-1 bg-blue-400 dark:bg-blue-600 rounded-lg hover:bg-blue-500 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 disabled:cursor-not-allowed"
      @click="comment"
    >
      Comment
    </button>
  </div>
</template>
