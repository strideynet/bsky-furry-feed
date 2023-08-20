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
    <shared-avatar class="mr-1" :url="data.avatar" :size="20" />
    {{ data.handle }}
  </nuxt-link>
</template>
