<script setup lang="ts">
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { AuditEvent } from "../../../proto/bff/v1/moderation_service_pb";

const $emit = defineEmits<{
  (event: "error", message: string): void;
}>();
const props = defineProps<{
  did: string;
  subject?: ProfileViewDetailed;
  hideCommentBox?: boolean;
}>();

const auditEvents: Ref<AuditEvent[]> = ref([]);

async function loadEvents() {
  $emit("error", "");

  const response = await api
    .listAuditEvents({
      filterSubjectDid: props.did,
    })
    .catch((err) => {
      $emit("error", err.rawMessage);
      return {
        auditEvents: [],
      };
    });
  auditEvents.value = response.auditEvents;
}

defineExpose({
  refresh() {
    loadEvents();
  },
});

async function comment(comment: string) {
  $emit("error", "");

  let ok = true;

  await api
    .createCommentAuditEvent({
      subjectDid: props.did,
      comment,
    })
    .catch((err) => {
      ok = false;
      $emit("error", err.rawMessage);
    });

  if (ok) await loadEvents();
}
const api = await useAPI();
await loadEvents();
</script>

<template>
  <h2 class="font-bold mb-3">Comments</h2>
  <action
    v-for="action in auditEvents.sort(
      (a, b) =>
        (a.createdAt?.toDate().getTime() || 0) -
        (b.createdAt?.toDate().getTime() || 0)
    )"
    :key="action.id"
    :action="action"
  />
  <p v-if="auditEvents.length === 0 && hideCommentBox" class="text-muted">
    No comments or audit events.
  </p>
  <shared-comment-box v-if="subject && !hideCommentBox" @comment="comment" />
</template>
