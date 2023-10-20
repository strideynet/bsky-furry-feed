<script lang="ts" setup>
import { Actor, ActorStatus } from "../../proto/bff/v1/types_pb";

const api = await useAPI();
const pending = ref(0);
const actor = ref<Actor>();
const error = ref<string>();

const nextActor = async () => {
  error.value = "";

  const queue = await api
    .listActors({ filterStatus: ActorStatus.PENDING })
    .catch((err) => {
      error.value = err.rawMessage;

      return {
        actors: [],
      };
    });

  const actors = queue.actors.filter(
    (a) => a.heldUntil && a.heldUntil.toDate() < new Date()
  );

  pending.value = actors.length - 1;
  const index = Math.floor(Math.random() * actors.length);
  actor.value = actors[index];
};

await nextActor();
</script>

<template>
  <shared-card v-if="error" variant="error">{{ error }}</shared-card>
  <user-profile
    v-else-if="actor"
    :did="actor.did"
    :pending="pending"
    variant="queue"
    @next="nextActor"
  />
  <shared-card v-else> No user is in the queue. </shared-card>
</template>
