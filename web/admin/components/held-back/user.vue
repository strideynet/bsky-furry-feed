<script setup lang="ts">
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { Actor } from "../../../proto/bff/v1/types_pb";
import { HoldBackPendingActorAuditPayload } from "../../../proto/bff/v1/moderation_service_pb";

const props = defineProps<{
  user: Actor & ProfileViewDetailed;
}>();

const api = await useAPI();
const holdBackCount = await api
  .listAuditEvents({
    filterSubjectDid: props.user.did,
  })
  .then(
    (r) =>
      r.auditEvents.filter((event) =>
        event.payload?.typeUrl.endsWith(
          "bff.v1.HoldBackPendingActorAuditPayload"
        )
      ).length
  );

const holdBackCountText = computed(() => {
  switch (holdBackCount) {
    case 1:
      return "once";
    case 2:
      return "twice";
    case 3:
      return "thrice";
    default:
      return `${holdBackCount} times`;
  }
});
</script>

<template>
  <div
    class="flex items-center gap-3 border border-gray-400 dark:border-gray-700 px-3 py-1.5"
  >
    <div class="flex md:items-center max-md:flex-col md:gap-3">
      <user-link class="text-truncate" :did="user.did" />
      <span v-if="user.displayName">{{ user.displayName }}</span>
      <div class="flex gap-2">
        <div class="text-sm max-md:text-xs">
          {{ user.postsCount || 0 }} <span class="text-muted">posts</span>
        </div>
        <div class="text-sm text-muted max-md:text-xs">
          held
          <span
            :class="{
              'text-yellow-400': holdBackCount === 2,
              'text-red-500': holdBackCount >= 3,
            }"
            >{{ holdBackCountText }}</span
          >
        </div>
        <div class="text-sm max-md:text-xs text-muted">
          back <shared-date :date="(user.heldUntil?.toDate() as Date)" />
        </div>
      </div>
    </div>

    <div class="ml-auto flex gap-1 items-center">
      <nuxt-link
        :href="`/users/${user.did}`"
        class="text-sm mr-1 py-0.5 max-md:py-1 max-md:px-3 px-2 rounded-lg border border-gray-400 dark:border-gray-700 hover:bg-slate-800"
      >
        View
      </nuxt-link>
      <div class="flex max-md:hidden">
        <user-queue-actions
          :did="user.did"
          :name="user.displayName || user.handle"
          small
          no-hold-back
        />
      </div>
    </div>
  </div>
</template>
