<script setup>
import { ActorStatus } from "../../proto/bff/v1/moderation_service_pb";

const api = await useAPI();
const queue = await api.listActors({ filterStatus: ActorStatus.PENDING });
const accept = () => {
  alert("todo: accepting");
};
const reject = () => {
  alert("todo: rejecting");
};
</script>

<template>
  <div class="max-w-[800px] py-4 px-3 mx-auto">
    <core-nav />
    <user-card :did="queue.actors[0]?.did" @accept="accept" @reject="reject" />
  </div>
</template>
