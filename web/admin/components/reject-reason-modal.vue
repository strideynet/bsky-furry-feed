<script setup lang="ts">
defineProps<{
  name: string;
}>();
defineEmits<{
  (event: "reject", reason: string): void;
  (event: "cancel"): void;
}>();

const reasons = [
  "Not a furry account",
  "Spam",
  "Hateful content",
  "Harassment",
  "Inappropriate sexual behavior",
  "AI-generated images",
];

const selectedReasons = ref(new Set<string>());
const showOther = ref(false);
const other = ref("");
const selectedReasonsText = computed(() =>
  [...selectedReasons.value, other.value]
    .filter(Boolean)
    .sort((a, b) => a.localeCompare(b))
    .join("; ")
);

function handleCheck(reason: string, value: boolean) {
  if (value) {
    selectedReasons.value.add(reason);
  } else {
    selectedReasons.value.delete(reason);
  }
}
</script>

<template>
  <div
    class="fixed flex top-0 left-0 right-0 bottom-0 z-10 justify-center items-center"
  >
    <shared-card class="dark:text-white z-10 bg-white dark:bg-gray-900">
      <h2 class="text-lg font-bold">Reject {{ name }}</h2>
      <p class="text-muted mb-3">
        Please select all applicable reasons for the rejection.
      </p>

      <ul class="mb-3">
        <li
          v-for="(reason, idx) in reasons"
          :key="idx"
          class="flex items-center gap-2 border border-gray-300 dark:border-gray-700 rounded-lg px-3 mb-1.5"
        >
          <input
            :id="`reason-${idx}`"
            type="checkbox"
            @input="
              handleCheck(reason, ($event.target as HTMLInputElement).checked)
            "
          />
          <label
            class="w-full h-full cursor-pointer py-1.5"
            :for="`reason-${idx}`"
          >
            {{ reason }}
          </label>
        </li>
        <li
          class="flex items-center gap-2 border border-gray-300 dark:border-gray-700 rounded-lg px-3 mb-1.5"
        >
          <input id="reason-other" v-model="showOther" type="checkbox" />
          <label
            id="other-label"
            class="w-full h-full cursor-pointer py-1.5"
            for="reason-other"
          >
            Other
          </label>
        </li>
        <li v-if="showOther">
          <input
            v-model="other"
            type="text"
            aria-labelledby="other-label"
            placeholder="Type a reason..."
            class="rounded-lg w-full py-1 px-3 bg-transparent border border-gray-300 dark:border-gray-700"
          />
        </li>
      </ul>

      <div class="flex justify-between">
        <button
          class="py-1 whitespace-nowrap px-2 text-white bg-gray-500 dark:bg-gray-600 hover:bg-gray-600 dark:hover:bg-gray-700 disabled:bg-gray-400 disabled:dark:bg-gray-500 rounded-lg disabled:cursor-not-allowed"
          @click="$emit('cancel')"
        >
          Cancel
        </button>

        <button
          class="py-1 px-2 mr-1 bg-red-500 dark:bg-red-600 hover:bg-red-600 dark:hover:bg-red-700 disabled:bg-red-400 disabled:dark:bg-red-500 rounded-lg disabled:cursor-not-allowed"
          :disabled="!selectedReasonsText"
          @click="$emit('reject', selectedReasonsText)"
        >
          Reject
        </button>
      </div>
    </shared-card>
    <div
      class="bg-black bg-opacity-50 absolute top-0 left-0 right-0 bottom-0"
      @click="$emit('cancel')"
    />
  </div>
</template>
