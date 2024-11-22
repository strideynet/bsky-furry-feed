<script lang="ts" setup>
const props = defineProps<{
  did?: string;
  hasAvatar: boolean;
  resize?: "72x72" | "20x20" | "webp";
  size: number;
  notRounded?: boolean;
}>();

const loading = ref(true);
const url = ref("");

function updateUrl() {
  loading.value = true;
  const img = new Image();
  let base = `https://bsky-cdn.codingpa.ws/avatar/${props.did}`;
  if (props.resize) {
    base += `/${props.resize}`;
  }
  img.addEventListener("load", () => {
    url.value = base;
    loading.value = false;
  });
  img.src = base;
}

updateUrl();

watch(() => props.did, updateUrl);
</script>

<template>
  <img
    v-if="did && hasAvatar && !loading"
    ref="img"
    :class="{ 'rounded-full': !notRounded }"
    :src="url"
    :height="size"
    :width="size"
    alt=""
  />
  <div
    v-else-if="loading && did && hasAvatar"
    class="loading-flash"
    :style="{ height: `${size}px`, width: `${size}px` }"
  ></div>
  <svg
    v-else
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="none"
  >
    <circle cx="12" cy="12" r="12" fill="#0070ff"></circle>
    <circle cx="12" cy="9.5" r="3.5" fill="#fff"></circle>
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      fill="#fff"
      d="M 12.058 22.784 C 9.422 22.784 7.007 21.836 5.137 20.262 C 5.667 17.988 8.534 16.25 11.99 16.25 C 15.494 16.25 18.391 18.036 18.864 20.357 C 17.01 21.874 14.64 22.784 12.058 22.784 Z"
    ></path>
  </svg>
</template>
