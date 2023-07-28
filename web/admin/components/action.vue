<script setup lang="ts">
import {
  ApprovalQueueAction,
  AuditEvent,
  CommentAuditPayload,
  ProcessApprovalQueueAuditPayload,
} from "../../proto/bff/v1/moderation_service_pb";

const props = defineProps<{
  action: AuditEvent;
}>();

const payload = computed(() => {
  if (!props.action.payload) return;

  const { typeUrl, value } = props.action.payload;

  switch (typeUrl.replace(/^type.googleapis.com\//, "")) {
    case "bff.v1.ProcessApprovalQueueAuditPayload":
      return ProcessApprovalQueueAuditPayload.fromBinary(value);
    case "bff.v1.CommentAuditPayload":
      return CommentAuditPayload.fromBinary(value);
  }
});

const type = computed(() => {
  const data = payload.value;

  if (data instanceof ProcessApprovalQueueAuditPayload) {
    return data.action === ApprovalQueueAction.APPROVE
      ? "queue_approval"
      : "queue_rejection";
  } else if (data instanceof CommentAuditPayload) {
    return "comment";
  }
});

const actionText = computed(() => {
  switch (type.value) {
    case "queue_approval":
      return "approved this user.";
    case "queue_rejection":
      return "rejected this user";
    case "comment":
      return "commented.";
  }
});

const comment = computed(() => {
  const data = payload.value;

  if (data instanceof CommentAuditPayload) {
    return data.comment;
  }
});
</script>

<template>
  <div class="flex items-start my-4">
    <div
      class="bg-gray-700 rounded-full flex items-center justify-center h-6 w-6 mr-3"
    >
      <icon-check v-if="type === 'queue_approval'" class="text-green-500" />
      <icon-cross v-else-if="type === 'queue_rejection'" class="text-red-500" />
      <icon-comment v-else-if="type === 'comment'" class="text-gray-300" />
    </div>
    <div class="flex-1">
      <div class="flex items-center gap-1">
        <user-link :did="action.actorDid" />
        {{ actionText }}
        <span class="text-gray-600 dark:text-gray-400 text-xs">{{
          action.createdAt?.toDate().toLocaleDateString("en", {
            day: "numeric",
            month: "short",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
          })
        }}</span>
      </div>
      <shared-card v-if="comment" no-padding class="text-sm px-3 py-2 mt-2">
        {{ comment }}
      </shared-card>
    </div>
  </div>
</template>
