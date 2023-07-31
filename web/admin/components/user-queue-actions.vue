<script setup lang="ts">
import { ApprovalQueueAction } from "../../proto/bff/v1/moderation_service_pb";

const props = defineProps<{
  did: string;
  handle: string;
  pending?: number;
}>();
const $emit = defineEmits(["loading", "next"]);

const loading = ref(false);

const accept = () => process(props.did, ApprovalQueueAction.APPROVE);
const reject = () => process(props.did, ApprovalQueueAction.REJECT);

const api = await useAPI();
async function process(did: string, action: ApprovalQueueAction) {
  loading.value = true;
  $emit("loading");
  await api.processApprovalQueue({
    did,
    action,
  });
  $emit("next");
  loading.value = false;
}
</script>

<template>
  <div
    class="mb-3 rounded-lg bg-orange-400 text-black px-3 py-1 flex items-baseline text-sm max-md:flex-col"
  >
    <span class="max-md:order-1"
      >{{ handle }} requested to be on the feed.</span
    >
    <span class="md:ml-auto max-md:w-full flex items-baseline">
      <span v-if="pending" class="text-xs text-gray-700 mx-1">
        ({{ pending }} more...)
      </span>
      <button
        class="py-0.5 px-2 max-md:ml-auto mr-1 text-white bg-blue-400 dark:bg-blue-600 rounded-lg hover:bg-blue-500 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 disabled:cursor-not-allowed"
        :disabled="loading"
        @click="accept"
      >
        Accept
      </button>

      <button
        class="py-0.5 px-2 bg-red-500 dark:bg-red-600 hover:bg-red-600 dark:hover:bg-red-700 disabled:bg-red-400 disabled:dark:bg-red-500 rounded-lg disabled:cursor-not-allowed"
        :disabled="loading"
        @click="reject"
      >
        Reject
      </button>
    </span>
  </div>
</template>
