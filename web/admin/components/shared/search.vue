<script setup>
import { search } from "~/lib/search";

const $emit = defineEmits(["toggleSearch"]);

const term = ref();

async function doSearch() {
  if (await search(term.value)) {
    term.value = "";
  }
}
</script>

<template>
  <div class="flex text-sm">
    <input
      v-model="term"
      class="py-1 px-2 rounded-l-lg border border-gray-300 text-black max-md:hidden"
      type="text"
      placeholder="User handle or did"
      @keydown="$event.key === 'Enter' ? doSearch() : null"
    />
    <button
      class="text-white bg-blue-500 hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 rounded-r-lg max-md:hidden px-1 py-1"
      @click="doSearch"
    >
      <icon-search />
    </button>
    <button
      class="text-white bg-blue-500 hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 rounded-lg md:hidden px-1 py-1"
      @click="$emit('toggleSearch')"
    >
      <icon-search />
    </button>
  </div>
</template>
