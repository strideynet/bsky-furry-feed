<script setup>
import { newAgent } from "~/lib/auth";

const agent = newAgent();
const { data } = await agent.getProfile({ actor: useRoute().params.did });
</script>

<template>
  <div>
    <user-card class="mb-5" :did="data.did" variant="profile" />
  </div>
</template>
