<script lang="ts" setup>
import { Actor, ActorStatus } from "../../proto/bff/v1/moderation_service_pb";

const api = await useAPI();
const pending = ref(0);
const actor = ref<Actor>();
const error = ref<{
  rawMessage: string;
}>(null);

const nextActor = async () => {
  const queue = await api.listActors({ filterStatus: ActorStatus.PENDING })
    .then((res) => {
      error.value = null;
      return res;
    })
    .catch((err) => {
      error.value = { rawMessage: err.rawMessage }

      return {
        actors: [],
      }
    });

  pending.value = queue.actors.length - 1;
  actor.value = queue.actors[0];
};

await nextActor();
</script>

<template>
  <shared-card v-if="error">{{ error.rawMessage }}</shared-card>
  <user-card v-else-if="actor" :did="actor.did" :pending="pending" variant="queue" @next="nextActor" />
  <shared-card v-else> No user is in the queue. </shared-card>
</template>
