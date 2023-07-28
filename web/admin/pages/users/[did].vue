<script setup lang="ts">
import { newAgent } from "~/lib/auth";
import { AuditEvent } from "../../../proto/bff/v1/moderation_service_pb";

const api = await useAPI();

const agent = newAgent();
const { data: subject } = await agent.getProfile({
  actor: String(useRoute().params.did),
});

const auditEvents: Ref<AuditEvent[]> = ref([]);

async function loadEvents() {
  const response = await api.listAuditEvents({ subjectDid: subject.did });
  auditEvents.value = response.auditEvents;
}

async function comment(comment: string) {
  await api.createCommentAuditEvent({
    subjectDid: subject.did,
    comment,
  });
  await loadEvents();
}

await loadEvents();
</script>

<template>
  <div>
    <user-card class="mb-5" :did="subject.did" variant="profile" />
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
    <shared-comment-box @comment="comment" />
  </div>
</template>
