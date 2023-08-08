<script setup lang="ts">
import {
  ApprovalQueueAction,
  AuditEvent,
  BanActorAuditPayload,
  CommentAuditPayload,
  ProcessApprovalQueueAuditPayload,
  UnapproveActorAuditPayload,
} from "../../proto/bff/v1/moderation_service_pb";

const props = defineProps<{
  action: AuditEvent;
  lookupUser?: boolean;
}>();

const payload = computed(() => {
  if (!props.action.payload) return;

  const { typeUrl, value } = props.action.payload;

  switch (typeUrl.replace(/^type.googleapis.com\//, "")) {
    case "bff.v1.ProcessApprovalQueueAuditPayload":
      return ProcessApprovalQueueAuditPayload.fromBinary(value);
    case "bff.v1.CommentAuditPayload":
      return CommentAuditPayload.fromBinary(value);
    case "bff.v1.UnapproveActorAuditPayload":
      return UnapproveActorAuditPayload.fromBinary(value);
    case "bff.v1.BanActorAuditPayload":
      return BanActorAuditPayload.fromBinary(value);
    default:
      console.warn(`Missing payload decoding: ${typeUrl}`);
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
  } else if (data instanceof UnapproveActorAuditPayload) {
    return "unapprove";
  } else if (data instanceof BanActorAuditPayload) {
    return "ban";
  }
});

const actionText = computed(() => {
  switch (type.value) {
    case "queue_approval":
      return props.lookupUser ? "approved" : "approved this user.";
    case "queue_rejection":
      return props.lookupUser ? "rejected" : "rejected this user.";
    case "comment":
      return props.lookupUser ? "commented on" : "commented.";
    case "unapprove":
      return props.lookupUser ? "unapproved" : "unapproved this user.";
    case "ban":
      return props.lookupUser ? "banned" : "banned this user.";
  }
});

const comment = computed(() => {
  const data = payload.value;

  if (data instanceof CommentAuditPayload) {
    return data.comment;
  } else if (data instanceof UnapproveActorAuditPayload) {
    return data.reason;
  } else if (data instanceof BanActorAuditPayload) {
    return data.reason;
  }
});
</script>

<template>
  <div class="flex items-start my-4">
    <div
      class="bg-gray-200 dark:bg-gray-800 rounded-full flex items-center justify-center h-7 w-7 mr-3 flex-shrink-0"
    >
      <icon-user-plus
        v-if="type === 'queue_approval'"
        class="text-gray-600 dark:text-gray-200"
      />
      <icon-user-minus
        v-else-if="type === 'queue_rejection' || type === 'unapprove'"
        class="text-gray-600 dark:text-gray-200"
      />
      <icon-slash v-else-if="type === 'ban'" class="text-red-500" />
      <icon-comment
        v-else-if="type === 'comment'"
        class="text-gray-600 dark:text-gray-200"
      />
    </div>
    <div class="flex-1">
      <div class="flex max-md:flex-wrap items-center gap-1">
        <user-link :did="action.actorDid" />
        {{ actionText }}
        <user-link v-if="lookupUser" :did="action.subjectDid" />
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
        <shared-markdown :markdown="comment" />
      </shared-card>
    </div>
  </div>
</template>
