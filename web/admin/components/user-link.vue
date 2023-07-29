<script setup lang="ts">
import { newAgent } from "~/lib/auth";

const props = defineProps<{ did: string }>();
const agent = newAgent();

const { data } = await agent.getProfile({
  actor: props.did,
});
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
