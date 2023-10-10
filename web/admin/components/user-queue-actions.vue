<script setup lang="ts">
import { showQueueActionConfirmation } from "~/lib/settings";
import { ApprovalQueueAction } from "../../proto/bff/v1/moderation_service_pb";
import { Duration } from "@bufbuild/protobuf";

const props = defineProps<{
  did: string;
  name: string;
  pending?: number;
}>();
const $emit = defineEmits(["loading", "next"]);

const loading = ref(false);

const accept = () => process(props.did, ApprovalQueueAction.APPROVE);
const reject = () => process(props.did, ApprovalQueueAction.REJECT);

const api = await useAPI();
async function process(did: string, action: ApprovalQueueAction) {
  if (showQueueActionConfirmation.value) {
    const verb = action === ApprovalQueueAction.APPROVE ? "approve" : "reject";

    if (!confirm(`Do you want to ${verb} ${props.name}?`)) {
      return;
    }
  }
  loading.value = true;
  $emit("loading");
  await api.processApprovalQueue({
    did,
    action,
  });
  $emit("next");
  loading.value = false;
}
async function holdBack() {
  if (showQueueActionConfirmation.value) {
    if (
      !confirm(
        `Do you want to hold back ${props.name}? Their account will appear in the queue again in 48 hours.`
      )
    ) {
      return;
    }
  }

  loading.value = true;
  $emit("loading");
  await api.holdBackPendingActor({
    did: props.did,
    duration: {
      seconds: BigInt(60 * 60 * 24 * 2),
    },
  });
  $emit("next");
  loading.value = false;
}
</script>

<template>
  <div
    class="mb-3 rounded-lg bg-orange-400 text-black px-3 py-1 flex items-baseline text-sm max-md:flex-col"
  >
    <span class="max-w-full flex max-md:mb-1"
      ><span class="truncate">{{ name }}</span>
      &nbsp;
      <span class="flex-shrink-0">requested to be on the feed.</span></span
    >
    <span
      class="md:ml-auto max-md:w-full flex items-baseline max-md:items-center"
    >
      <span v-if="pending" class="text-xs text-gray-700 mx-1">
        ({{ pending }} more...)
      </span>
      <button
        class="py-0.5 max-md:py-1 max-md:px-3 px-2 max-md:ml-auto mr-1 text-white bg-blue-500 dark:bg-blue-600 rounded-lg hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 disabled:cursor-not-allowed"
        :disabled="loading"
        @click="accept"
      >
        Accept
      </button>

      <button
        class="py-0.5 max-md:py-1 max-md:px-3 px-2 mr-1 bg-red-500 dark:bg-red-600 hover:bg-red-600 dark:hover:bg-red-700 disabled:bg-red-400 disabled:dark:bg-red-500 rounded-lg disabled:cursor-not-allowed"
        :disabled="loading"
        @click="reject"
      >
        Reject
      </button>

      <button
        v-if="false /* waiting for backend deployment! */"
        class="py-0.5 max-md:py-1 whitespace-nowrap max-md:px-3 px-2 text-white bg-gray-500 dark:bg-gray-600 hover:bg-gray-600 dark:hover:bg-gray-700 disabled:bg-gray-400 disabled:dark:bg-gray-500 rounded-lg disabled:cursor-not-allowed"
        :disabled="loading"
        @click="holdBack"
      >
        Hold back
      </button>
    </span>
  </div>
</template>
