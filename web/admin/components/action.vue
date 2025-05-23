<script setup lang="ts">
import {
  ApprovalQueueAction,
  AuditEvent,
  BanActorAuditPayload,
  CommentAuditPayload,
  CreateActorAuditPayload,
  ForceApproveActorAuditPayload,
  ProcessApprovalQueueAuditPayload,
  UnapproveActorAuditPayload,
  HoldBackPendingActorAuditPayload,
  AssignRolesAuditPayload,
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
    case "bff.v1.CreateActorAuditPayload":
      return CreateActorAuditPayload.fromBinary(value);
    case "bff.v1.ForceApproveActorAuditPayload":
      return ForceApproveActorAuditPayload.fromBinary(value);
    case "bff.v1.HoldBackPendingActorAuditPayload":
      return HoldBackPendingActorAuditPayload.fromBinary(value);
    case "bff.v1.AssignRolesAuditPayload":
      return AssignRolesAuditPayload.fromBinary(value);
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
  } else if (data instanceof CreateActorAuditPayload) {
    return "track";
  } else if (data instanceof ForceApproveActorAuditPayload) {
    return "force_approve";
  } else if (data instanceof HoldBackPendingActorAuditPayload) {
    return "hold_back";
  } else if (data instanceof AssignRolesAuditPayload) {
    return "assign_roles";
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
    case "track":
      return props.lookupUser
        ? "started tracking"
        : "started tracking this user.";
    case "force_approve":
      return props.lookupUser ? "force-approved" : "force-approved this user.";
    case "hold_back":
      return props.lookupUser ? "held back" : "held back this user.";
    case "assign_roles":
      return props.lookupUser
        ? "assigned roles"
        : "assigned roles to this user.";
  }
});

const comment = computed(() => {
  const data = payload.value;

  if (data instanceof CommentAuditPayload) {
    return data.comment;
  } else if (data instanceof AssignRolesAuditPayload) {
    const mapRoles = (roles: string[]) =>
      roles.map((r) => r.replace(/^./, (r) => r.toUpperCase())).join(", ") ||
      "<span class='text-muted'>None</span>";

    return `**Before**: ${mapRoles(data.rolesBefore)}\n\n**After**: ${mapRoles(
      data.rolesAfter
    )}`;
  } else if (
    data instanceof UnapproveActorAuditPayload ||
    data instanceof BanActorAuditPayload ||
    data instanceof CreateActorAuditPayload ||
    data instanceof ForceApproveActorAuditPayload ||
    data instanceof ProcessApprovalQueueAuditPayload
  ) {
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
        v-if="type === 'queue_approval' || type === 'force_approve'"
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
      <icon-database
        v-else-if="type === 'track'"
        class="text-gray-600 dark:text-gray-200"
      />
      <icon-clock
        v-else-if="type === 'hold_back'"
        class="text-gray-600 dark:text-gray-200"
      />
      <icon-key
        v-else-if="type === 'assign_roles'"
        class="text-gray-600 dark:text-gray-200"
      />
    </div>
    <div class="flex-1">
      <div class="flex max-md:flex-wrap items-center gap-1">
        <user-link :did="action.actorDid" />
        {{ actionText }}
        <user-link v-if="lookupUser" :did="action.subjectDid" />
        <shared-date
          v-if="action.createdAt"
          class="text-muted text-xs"
          :date="action.createdAt?.toDate()"
        />
      </div>
      <shared-card v-if="comment" no-padding class="text-sm px-3 py-2 mt-2">
        <shared-markdown :markdown="comment" />
      </shared-card>
    </div>
  </div>
</template>
