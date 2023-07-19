<script setup>
import * as auth from "~/lib/auth";
useHead({
  title: "furryli.st Admin",
  meta: [{ name: "robots", content: "noindex" }],
  link: [
    { rel: "preconnect", href: "https://fonts.googleapis.com" },
    { rel: "preconnect", href: "https://fonts.gstatic.com" },
    {
      href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;700&display=swap",
      rel: "stylesheet",
    },
  ],
  bodyAttrs: {
    class: "bg-white dark:bg-gray-900 dark:text-white",
  },
});

const identifier = ref();
const password = ref();
const error = ref();

const user = await useUser();

async function login() {
  error.value = null;
  const isSignedIn = await auth
    .login(identifier.value, password.value)
    .catch((error) => ({ error }));

  error.value = !isSignedIn;
}
</script>

<template>
  <NuxtPage v-if="user" />
  <div class="flex items-center justify-center fixed w-full h-full" v-else>
    <div
      class="mx-auto bg-gray-50 border border-gray-400 dark:border-gray-700 dark:bg-gray-800 py-4 px-5 rounded-lg w-[400px] max-w-[80vw]"
    >
      <h1 class="text-3xl font-bold mb-4">Login</h1>

      <div class="flex flex-col mb-4">
        <label class="font-bold mb-1" for="name">Handle</label>
        <input
          class="bg-white dark:bg-gray-900 rounded border border-gray-400 dark:border-gray-700 px-2 py-1"
          id="name"
          type="text"
          v-model="identifier"
        />
      </div>

      <div class="flex flex-col mb-4">
        <label class="font-bold mb-1" for="password">App password</label>
        <input
          class="bg-white dark:bg-gray-900 rounded border border-gray-400 dark:border-gray-700 px-2 py-1"
          id="password"
          type="password"
          v-model="password"
        />
      </div>

      <div class="flex">
        <button class="ml-auto px-3 py-2 rounded-lg bg-blue-600" @click="login">
          Login
        </button>
      </div>
    </div>
  </div>
</template>
