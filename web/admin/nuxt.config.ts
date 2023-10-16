// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ["@nuxtjs/tailwindcss"],
  css: ["~/assets/css/main.css"],
  devtools: { enabled: true },
  ssr: false,
  runtimeConfig: {
    public: {
      apiUrl: process.env.ADMIN_API_URL || "https://feed.furryli.st",
    },
  },
});
