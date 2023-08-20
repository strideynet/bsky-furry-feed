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

  if (isSignedIn.error)
    error.value = {
      message: isSignedIn.error
    };
}
</script>

<template>
  <div v-if="user" class="max-w-[800px] py-4 px-3 mx-auto">
    <core-nav />
    <NuxtPage />
  </div>
  <div v-else class="flex items-center justify-center fixed w-full h-full">
    <div
      class="mx-auto bg-gray-50 border border-gray-400 dark:border-gray-700 dark:bg-gray-800 py-4 px-5 rounded-lg w-[400px] max-w-[80vw]">
      <h1 class="text-3xl font-bold mb-4">Login</h1>

      <div class="flex flex-col mb-4">
        <label class="font-bold mb-1" for="name">Handle</label>
        <input id="name" v-model="identifier"
          class="bg-white dark:bg-gray-900 rounded border border-gray-400 dark:border-gray-700 px-2 py-1" type="text" />
      </div>

      <div class="flex flex-col mb-4">
        <label class="font-bold mb-1" for="password">App password</label>
        <input id="password" v-model="password"
          class="bg-white dark:bg-gray-900 rounded border border-gray-400 dark:border-gray-700 px-2 py-1"
          type="password" />
      </div>

      <div class="flex">
        <label v-if="error" class="mr-auto px-1 py-2 text-red-600">
          {{ error.message }}
        </label>
        <button class="ml-auto px-3 py-2 rounded-lg bg-blue-600" @click="login">
          Login
        </button>
      </div>
    </div>
  </div>
</template>
