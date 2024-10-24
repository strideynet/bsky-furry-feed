<script setup lang="ts">
import { newAgent } from "~/lib/auth";
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";

const props = defineProps<{
  did: string;
  pending?: number;
  variant: "profile" | "queue";
}>();
const $emit = defineEmits<{
  (event: "next"): void;
}>();

const error = ref<string>();
const auditLog = ref() as Ref<{ refresh(): Promise<void> }>;

const subject = ref<ProfileViewDetailed>();
const agent = newAgent();
async function loadProfile() {
  const { data } = await agent
    .getProfile({
      actor: props.did,
    })
    .catch(() => {
      return { data: undefined };
    });
  subject.value = data;
}

async function refresh() {
  await loadProfile();
  await auditLog.value?.refresh();
}

function handleNext() {
  $emit("next");
  if (props.variant === "profile") {
    refresh();
  }
}

watch(
  () => props.did,
  () => refresh()
);

await refresh();
</script>

<template>
  <div>
    <user-card
      class="mb-5"
      :did="subject?.did || props.did"
      :pending="pending"
      :variant="variant"
      @next="handleNext"
    />
    <shared-card v-if="error" variant="error">{{ error }}</shared-card>
    <user-audit-log
      v-else
      ref="auditLog"
      :subject="subject"
      :did="subject?.did || props.did"
      @error="error = $event"
    />
  </div>
</template>
