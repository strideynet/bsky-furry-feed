<script setup>
import { newAgent } from "~/lib/auth";

const term = ref();

const search = async () => {
  const agent = newAgent();
  const { data, success } = await agent
    .getProfile({ actor: term.value })
    .catch(() => ({ success: false }));
  if (!success) {
    alert("Could not find user. Please check handle or did, and try again.");
    return;
  }
  useRouter().push(`/users/${data.did}`);
  term.value = "";
};
</script>

<template>
  <div class="flex">
    <input
      v-model="term"
      class="py-1 px-2 rounded-l-lg border border-gray-300 text-black"
      type="text"
      placeholder="User handle or did"
      @keydown="$event.key === 'Enter' ? search() : null"
    />
    <button
      class="text-white bg-blue-400 dark:bg-blue-500 rounded-r-lg px-1"
      @click="search"
    >
      <icon-search />
    </button>
  </div>
</template>
