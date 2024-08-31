<script setup lang="ts">
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { getProfile } from "~/lib/cached-bsky";
import { Actor } from "../../../proto/bff/v1/types_pb";
import { addSISuffix } from "~/lib/util";

const props = defineProps<{
  did: string;
}>();

const loading = ref(true);
const profile = ref() as Ref<ProfileViewDetailed>;
const actor = ref<Actor>();

onMounted(async () => {
  profile.value = await getProfile(props.did);
  const api = await useAPI();
  const resp = await api.getActor({ did: props.did });
  actor.value = resp.actor;
  loading.value = false;
});
</script>

<template>
  <shared-card no-padding class="absolute bottom-2.5 w-[200px] bg-slate-800">
    <div v-if="loading" class="py-1.5 px-2">Loading...</div>
    <template v-else>
      <div
        class="flex items-center py-1.5 px-2 border-b border-gray-300 dark:border-gray-700 gap-2"
      >
        <shared-avatar
          :did="did"
          :size="48"
          resize="72x72"
          has-avatar
          not-rounded
          class="rounded-lg"
        />
        <div>
          <div class="truncate">{{ profile.displayName }}</div>
          <div class="text-muted truncate">@{{ profile.handle }}</div>
        </div>
      </div>
      <div class="px-2 py-1.5 flex items-center">
        <user-status-badge :status="actor?.status" class="text-xs" />
        <span class="ml-auto text-xs">
          {{ addSISuffix(profile.followersCount || 0) }} followers
        </span>
      </div>
    </template>
  </shared-card>
</template>
