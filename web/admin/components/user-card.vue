<script lang="ts" setup>
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { newAgent } from "~/lib/auth";

const props = defineProps<{ did: string; loading: boolean }>();
defineEmits(["accept", "reject"]);

const agent = newAgent();
const data = ref<ProfileViewDetailed>();
const loadProfile = async () => {
  const result = await agent.getProfile({
    actor: props.did,
  });
  data.value = result.data;
};

watch(
  () => props.did,
  () => loadProfile()
);

await loadProfile();
</script>

<template>
  <shared-card v-if="data">
    <div class="flex gap-3 items-center">
      <shared-avatar :url="data.avatar" />
      <div class="flex flex-col">
        <div class="text-lg">{{ data.displayName || data.handle }}</div>
        <div class="meta">
          <span class="meta-item">
            <nuxt-link
              class="underline hover:no-underline text-gray-600 dark:text-gray-400"
              :href="`https://bsky.app/profile/${data.handle}`"
              target="_blank"
              >@{{ data.handle }}</nuxt-link
            >
          </span>
          <span class="meta-item">
            {{ data.followersCount }}
            <span class="text-gray-600 dark:text-gray-400">followers</span>
          </span>
          <span class="meta-item">
            {{ data.followsCount }}
            <span class="text-gray-600 dark:text-gray-400">follows</span>
          </span>
          <span class="meta-item">
            {{ data.postsCount }}
            <span class="text-gray-600 dark:text-gray-400">posts</span>
          </span>
        </div>
      </div>
    </div>
    <div class="mt-5">
      <shared-bsky-description
        v-if="data.description"
        :description="data.description"
      />
    </div>
    <div class="flex gap-3 mt-5">
      <button
        class="px-3 py-2 bg-blue-400 dark:bg-blue-500 rounded-lg"
        :disabled="loading"
        @click="$emit('accept')"
      >
        Accept
      </button>
      <button
        class="px-3 py-2 bg-red-500 dark:bg-red-600 rounded-lg"
        :disabled="loading"
        @click="$emit('reject')"
      >
        Reject
      </button>
    </div>
  </shared-card>
  <shared-card class="bg-red-200 dark:bg-red-700" v-else>
    Profile with did {{ did }} was not found.
  </shared-card>
</template>

<style scoped>
.meta .meta-item:not(:last-child)::after {
  content: " · ";
  text-decoration: none;
}
</style>
