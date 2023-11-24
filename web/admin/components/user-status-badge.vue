<script setup lang="ts">
import { ActorStatus } from "../../proto/bff/v1/types_pb";
import { ACTOR_STATUS_LABELS } from "~/lib/constants";

const props = defineProps<{ status?: ActorStatus }>();

const statusLabel = computed(() =>
  props.status === undefined ? "Untracked" : ACTOR_STATUS_LABELS[props.status]
);
const background = computed(() =>
  props.status === undefined
    ? "bg-gray-800 text-gray-400 text-white"
    : {
        [ActorStatus.UNSPECIFIED]: "bg-gray-600 text-white",
        [ActorStatus.PENDING]: "bg-orange-400 text-black",
        [ActorStatus.APPROVED]: "bg-green-700 text-white",
        [ActorStatus.BANNED]: "bg-red-700 text-white",
        [ActorStatus.NONE]: "bg-zinc-700 text-white",
      }[props.status]
);
</script>

<template>
  <span class="rounded-full py-0.5 px-2" :class="background">
    {{ statusLabel }}
  </span>
</template>
