<script setup lang="ts">
import { showQueueActionConfirmation } from "~/lib/settings";
import { ApprovalQueueAction } from "../../../proto/bff/v1/moderation_service_pb";

const showRejectModal = ref(false);
const loading = ref(false);
const props = defineProps<{
  did: string;
  name: string;
  rejectOnly?: boolean;
  noHoldBack?: boolean;
  small?: boolean;
}>();
const $emit = defineEmits(["loading", "next"]);

const accept = () => process(props.did, ApprovalQueueAction.APPROVE);
const reject = (reason: string) =>
  process(props.did, ApprovalQueueAction.REJECT, reason);

const api = await useAPI();

async function process(
  did: string,
  action: ApprovalQueueAction,
  reason?: string
) {
  if (showQueueActionConfirmation.value && !reason) {
    const verb = action === ApprovalQueueAction.APPROVE ? "approve" : "reject";

    if (!confirm(`Do you want to ${verb} ${props.name}?`)) {
      return;
    }
  }
  loading.value = true;
  $emit("loading");
  try {
    await api.processApprovalQueue({
      did,
      action,
      reason,
    });
  } catch (err) {
    const message = String(err);
    if (!message.includes("candidate actor status was")) {
      alert("An error occurred while processing the queue: " + message);
    }
  }
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
  try {
    const days = 60 * 60 * 24;
    await api.holdBackPendingActor({
      did: props.did,
      duration: {
        seconds: BigInt(days * 7),
      },
    });
  } catch (err) {
    const message = String(err);
    if (!message.includes("candidate actor status was")) {
      alert("An error occurred while holding back the user: " + message);
    }
  }
  $emit("next");
  loading.value = false;
}

function promptReject() {
  if (props.rejectOnly) {
    reject("Profile no longer exists");
    return;
  }
  showRejectModal.value = true;
}
</script>

<template>
  <reject-reason-modal
    v-if="showRejectModal"
    :name="name"
    @cancel="showRejectModal = false"
    @reject="
      (reason) => {
        showRejectModal = false;
        reject(reason);
      }
    "
  />
  <button
    v-if="!rejectOnly"
    class="py-0.5 max-md:py-1 max-md:px-3 px-2 max-md:ml-auto mr-1 text-white bg-blue-500 dark:bg-blue-600 rounded-lg hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 disabled:cursor-not-allowed"
    :class="{ 'text-sm': small }"
    :disabled="loading"
    @click="accept"
  >
    Accept
  </button>

  <button
    :class="{ 'text-sm': small }"
    class="py-0.5 max-md:py-1 max-md:px-3 px-2 mr-1 bg-red-500 dark:bg-red-600 hover:bg-red-600 dark:hover:bg-red-700 disabled:bg-red-400 disabled:dark:bg-red-500 rounded-lg disabled:cursor-not-allowed"
    :disabled="loading"
    @click="promptReject"
  >
    Reject
  </button>

  <button
    v-if="!rejectOnly && !noHoldBack"
    :class="{ 'text-sm': small }"
    class="py-0.5 max-md:py-1 whitespace-nowrap max-md:px-3 px-2 text-white bg-gray-500 dark:bg-gray-600 hover:bg-gray-600 dark:hover:bg-gray-700 disabled:bg-gray-400 disabled:dark:bg-gray-500 rounded-lg disabled:cursor-not-allowed"
    :disabled="loading"
    @click="holdBack"
  >
    Hold back
  </button>
</template>
