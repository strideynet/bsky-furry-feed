<script lang="ts" setup>
import {
  Actor,
  ActorStatus,
  ApprovalQueueAction,
} from "../../proto/bff/v1/moderation_service_pb";

const api = await useAPI();
const pending = ref(0);
const actor = ref<Actor>();
const loading = ref(false);

const nextActor = async () => {
  const queue = await api.listActors({ filterStatus: ActorStatus.PENDING });
  pending.value = queue.actors.length - 1;
  actor.value = queue.actors[0];
};

const accept = () =>
  process(actor.value?.did as string, ApprovalQueueAction.APPROVE);
const reject = () =>
  process(actor.value?.did as string, ApprovalQueueAction.REJECT);
const process = async (did: string, action: ApprovalQueueAction) => {
  loading.value = true;
  await api.processApprovalQueue({
    did,
    action,
  });
  await nextActor();
  loading.value = false;
};

await nextActor();
</script>

<template>
  <user-card
    v-if="actor"
    :did="actor.did"
    :loading="loading"
    :pending="pending"
    variant="queue"
    @accept="accept"
    @reject="reject"
  />
  <shared-card v-else> No user is in the queue. </shared-card>
</template>
