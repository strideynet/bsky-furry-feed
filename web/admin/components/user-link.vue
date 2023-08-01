<script setup lang="ts">
import { getProfile } from "~/lib/cached-bsky";

const props = defineProps<{ did: string }>();

const { data } = await getProfile(props.did);
</script>

<template>
  <nuxt-link
    class="flex items-center underline hover:no-underline"
    :href="`/users/${data.did}`"
  >
    <img
      class="rounded-full mr-1"
      :src="data.avatar"
      height="20"
      width="20"
      alt=""
    />
    {{ data.handle }}
  </nuxt-link>
</template>
