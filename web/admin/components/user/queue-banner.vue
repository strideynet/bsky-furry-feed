<script setup lang="ts">
defineProps<{
  did: string;
  name: string;
  pending?: number;
  rejectOnly?: boolean;
}>();
defineEmits(["loading", "next"]);
</script>

<template>
  <div
    class="mb-3 rounded-lg bg-orange-400 text-black px-3 py-1 flex items-baseline text-sm max-md:flex-col"
  >
    <span class="max-w-full flex max-md:mb-1"
      ><span class="truncate">{{ name }}</span>
      &nbsp;
      <span class="flex-shrink-0">requested to be on the feed.</span></span
    >
    <span
      class="md:ml-auto max-md:w-full flex items-baseline max-md:items-center"
    >
      <span v-if="pending" class="text-xs text-gray-700 mx-1">
        ({{ pending }} more...)
      </span>
      <user-queue-actions
        :did="did"
        :name="name"
        :reject-only="rejectOnly"
        @loading="$emit('loading')"
        @next="$emit('next')"
      />
    </span>
  </div>
</template>
