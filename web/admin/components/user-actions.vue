<script setup lang="ts">
import { ActorStatus } from "../../proto/bff/v1/moderation_service_pb";

const props = defineProps<{ did: string; status: ActorStatus }>();
const $emit = defineEmits(["update"]);
const api = await useAPI();

async function remove() {
  if (!confirm("Are you sure to remove this user from the list?")) {
    return;
  }

  const reason = prompt("Enter the reason for unapproving the user.");

  if (!reason) {
    alert("A reason is required to unapprove the user.");
    return;
  }

  await api.unapproveActor({ actorDid: props.did, reason });
  $emit("update");
}

async function ban() {
  if (!confirm("Are you sure to ban this user from the list?")) {
    return;
  }

  const reason = prompt("Enter the reason for banning the user.");

  if (!reason) {
    alert("A reason is required to ban the user.");
    return;
  }

  await api.banActor({ actorDid: props.did, reason });
  $emit("update");
}
</script>

<template>
  <div class="mb-3 flex items-center text-sm gap-2">
    <span>
      <user-status-badge :status="status" />
    </span>
    <span class="ml-auto">&nbsp;</span>
    <span v-if="props.status !== ActorStatus.BANNED">
      <button class="rounded-lg py-1 px-2 bg-red-700 text-white" @click="ban">
        Ban user
      </button>
    </span>
    <span v-if="props.status === ActorStatus.APPROVED">
      <button
        class="rounded-lg py-1 px-2 bg-zinc-700 text-white"
        @click="remove"
      >
        Remove from list
      </button>
    </span>
  </div>
</template>
