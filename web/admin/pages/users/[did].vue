<script setup lang="ts">
import { newAgent } from "~/lib/auth";
import { AuditEvent } from "../../../proto/bff/v1/moderation_service_pb";
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";

const api = await useAPI();

const error = ref<string>();

const subject = ref() as Ref<ProfileViewDetailed>;
const agent = newAgent();
async function loadProfile() {
  const { data } = await agent.getProfile({
    actor: String(useRoute().params.did),
  });
  subject.value = data;
}

const auditEvents: Ref<AuditEvent[]> = ref([]);

async function loadEvents() {
  error.value = "";

  const response = await api
    .listAuditEvents({
      filterSubjectDid: subject.value.did,
    })
    .catch((err) => {
      error.value = err.rawMessage;
      return {
        auditEvents: [],
      };
    });
  auditEvents.value = response.auditEvents;
}

async function comment(comment: string) {
  error.value = "";

  await api
    .createCommentAuditEvent({
      subjectDid: subject.value.did,
      comment,
    })
    .catch((err) => {
      error.value = err.rawMessage;
    });

  if (!error.value) await loadEvents();
}

async function refresh() {
  await loadProfile();
  await loadEvents();
}

await refresh();
</script>

<template>
  <div>
    <shared-card v-if="error" variant="error">{{ error }}</shared-card>
    <div v-else>
      <user-card
        class="mb-5"
        :did="subject.did"
        variant="profile"
        @next="refresh"
      />
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
  </div>
</template>
