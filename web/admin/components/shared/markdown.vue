<script setup lang="ts">
import DOMPurify from "dompurify";
import { marked } from "marked";

const props = defineProps<{
  markdown: string;
}>();

const renderedMarkdown = computed(() => {
  const rendered = marked.parse(props.markdown, {
    gfm: true,
  });
  return DOMPurify.sanitize(rendered);
});
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div class="markdown" v-html="renderedMarkdown" />
</template>

<style scoped>
.markdown >>> a {
  @apply underline;
  @apply text-blue-600;
  @apply dark:text-blue-400;
}

.markdown >>> a:hover {
  @apply no-underline;
}

.markdown >>> img,
.markdown >>> table {
  @apply hidden;
}

.markdown >>> blockquote {
  @apply pl-2 border-l-4 border-gray-400 dark:border-gray-700;
}

.markdown >>> p:not(:last-child),
.markdown >>> blockquote:not(:last-child) {
  @apply mb-2;
}
</style>
