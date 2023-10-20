<script setup lang="ts">
const props = defineProps<{ date: Date }>();

const generate = () => {
  const diff = Math.round((Date.now() - props.date.getTime()) / 1000);
  const isFuture = diff < 0;
  const delta = Math.abs(diff);

  const duration = durationToText(delta);

  return `${isFuture ? "in " : ""}${duration}${isFuture ? "" : " ago"}`;
};

const durationToText = (duration: number) => {
  let text = "";

  if (duration < 30) {
    text = "a few seconds";
  } else if (duration < 60) {
    text = "a minute";
  } else if (duration / 60 < 3) {
    text = "a few minutes";
  } else if (duration / 60 < 60) {
    text = `${Math.round(duration / 60)} minutes`;
  } else if (duration / 60 / 60 < 24) {
    text = `${Math.round(duration / 60 / 60)} hours`;
  } else {
    const days = Math.round(duration / 60 / 60 / 24);
    text = `${days} day${days === 1 ? "" : "s"}`;
  }

  return text;
};

const relativeDate = ref(generate());
let interval: any;

onMounted(() => {
  interval = setInterval(() => (relativeDate.value = generate()), 1000 * 3);
});

onBeforeUnmount(() => clearInterval(interval));
</script>

<template>
  <time
    :title="
      date.toLocaleDateString('en', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    "
    >{{ relativeDate }}</time
  >
</template>
