<script setup>
import { logout, newAgent } from "~/lib/auth";

const user = await useUser();
const agent = newAgent();
const profile = await agent.getProfile({ actor: user.value.did });
</script>

<template>
  <nav
    class="flex items-center gap-2 border border-gray-300 dark:border-gray-700 rounded-lg px-4 py-3 mb-5"
  >
    <nuxt-link href="/" class="mr-1">
      <img
        class="rounded-lg"
        src="/icon-32.webp"
        height="32"
        width="32"
        alt=""
      />
    </nuxt-link>
    <h1 class="text-xl font-bold">Admin</h1>
    <div class="ml-auto flex items-center gap-2">
      <img
        class="rounded-full"
        :src="profile.data.avatar"
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
</template>
