<script setup lang="ts">
import { ActorStatus } from "../../proto/bff/v1/types_pb";

const props = defineProps<{ did: string; status?: ActorStatus }>();
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

async function createActor() {
  const reason = prompt("Enter the reason for tracking the user.");

  if (!reason) {
    alert("A reason is required to track the user.");
    return;
  }

  await api.createActor({ actorDid: props.did, reason });
  $emit("update");
}

async function forceApprove() {
  const reason = prompt("Enter the reason for force-approving the user.");

  if (!reason) {
    alert("A reason is required to force-approve the user.");
    return;
  }

  await api.forceApproveActor({ actorDid: props.did, reason });
  $emit("update");
}
</script>

<template>
  <div class="mb-3 flex items-center text-sm gap-2">
    <span>
      <user-status-badge :status="status" />
    </span>
    <span class="ml-auto">&nbsp;</span>
    <span v-if="props.status === undefined">
      <button
        class="rounded-lg py-1 px-2 bg-gray-200 text-black dark:bg-gray-600 dark:text-white"
        @click="createActor"
      >
        Track user
      </button>
    </span>
    <span
      v-if="props.status !== undefined && props.status === ActorStatus.NONE"
    >
      <button
        class="rounded-lg py-1 px-2 text-white bg-blue-500 dark:bg-blue-600 hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500"
        @click="forceApprove"
      >
        Force-approve
      </button>
    </span>
    <span
      v-if="props.status !== undefined && props.status !== ActorStatus.BANNED"
    >
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
