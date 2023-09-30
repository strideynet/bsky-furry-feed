<script setup lang="ts">
import { getProfile } from "~/lib/cached-bsky";

const props = defineProps<{ did: string }>();

const profile = await getProfile(props.did);
</script>

<template>
  <nuxt-link
    class="flex items-center underline hover:no-underline"
    :href="`/users/${profile.did}`"
  >
    <shared-avatar class="mr-1" :url="profile.avatar" :size="20" />
    {{ profile.handle }}
  </nuxt-link>
</template>
