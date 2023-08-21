<script setup>
import { search } from "~/lib/search";
import { logout } from "~/lib/auth";

const profile = await useProfile();
const showSearch = ref(false);

const term = ref("");
async function doSearch() {
  if (await search(term.value)) {
    term.value = "";
    showSearch.value = false;
  }
}
</script>

<template>
  <nav
    class="flex items-center gap-2 border border-gray-300 dark:border-gray-700 rounded-lg px-4 py-3 mb-5"
  >
    <nuxt-link href="/" class="mr-1" aria-label="Home">
      <img
        class="rounded-lg"
        src="/icon-32.webp"
        height="32"
        width="32"
        alt=""
      />
    </nuxt-link>

    <nuxt-link class="nav-link" href="/"> Queue </nuxt-link>

    <nuxt-link class="nav-link" href="/audit-log"> Audit log </nuxt-link>

    <div class="ml-auto flex items-center gap-2">
      <shared-search @toggle-search="showSearch = !showSearch" />
      <img
        class="rounded-full"
        :src="profile.avatar"
        height="32"
        width="32"
        alt=""
      />
      <button
        class="text-white bg-gray-700 px-2 py-1 rounded-lg"
        @click="logout"
      >
        Logout
      </button>
    </div>
  </nav>
  <div
    v-if="showSearch"
    class="flex text-sm px-4 py-3 mb-5 border border-gray-300 dark:border-gray-700 rounded-lg"
  >
    <input
      v-model="term"
      class="py-1 px-2 w-full rounded-l-lg border border-gray-300 text-black"
      type="text"
      placeholder="User handle or did"
      @keydown="$event.key === 'Enter' ? doSearch() : null"
    />
    <button
      class="text-white hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 rounded-r-lg px-1 py-1"
      @click="doSearch"
    >
      <icon-search />
    </button>
  </div>
</template>

<style scoped>
.nav-link {
  @apply px-2 py-1 text-sm rounded-lg;
}

.nav-link.router-link-active,
.nav-link:hover {
  @apply bg-slate-600 bg-opacity-50;
}
</style>
