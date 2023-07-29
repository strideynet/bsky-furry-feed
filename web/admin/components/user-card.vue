<script lang="ts" setup>
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { newAgent } from "~/lib/auth";
import { ActorStatus } from "../../proto/bff/v1/moderation_service_pb";

const props = defineProps<{
  did: string;
  loading?: boolean;
  pending?: number;
  variant: "queue" | "profile";
}>();
defineEmits(["accept", "reject"]);

const api = await useAPI();
const isArtist = ref(false);
const status = ref<ActorStatus | undefined>();
const agent = newAgent();
const data = ref<ProfileViewDetailed>();
const loadProfile = async () => {
  const result = await agent.getProfile({
    actor: props.did,
  });
  data.value = result.data;
  const { actor } = await api
    .getActor({ did: result.data.did })
    .catch(() => ({ actor: undefined }));
  isArtist.value = Boolean(actor?.isArtist);
  status.value = actor?.status;
};

watch(
  () => props.did,
  () => loadProfile()
);

await loadProfile();
</script>

<template>
  <shared-card v-if="data" :class="{ loading }">
    <div
      v-if="variant === 'queue' && status === ActorStatus.PENDING"
      class="flex items-center gap-3 pt-2 mb-5"
    >
      <button
        class="px-3 py-2 bg-blue-400 dark:bg-blue-600 rounded-lg hover:bg-blue-500 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 disabled:cursor-not-allowed"
        :disabled="loading"
        @click="$emit('accept')"
      >
        Accept
      </button>
      <button
        class="px-3 py-2 bg-red-500 dark:bg-red-600 hover:bg-red-600 dark:hover:bg-red-700 disabled:bg-red-400 disabled:dark:bg-red-500 rounded-lg disabled:cursor-not-allowed"
        :disabled="loading"
        @click="$emit('reject')"
      >
        Reject
      </button>

      <span
        v-if="pending > 0"
        class="ml-auto text-sm text-gray-600 dark:text-gray-400"
        >and {{ pending }} more...</span
      >
    </div>

    <div class="flex gap-3 items-center mb-5">
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
        <div v-if="variant === 'profile'" class="meta">
          <span class="meta-item inline-flex items-center">
            <icon-check
              v-if="status === ActorStatus.APPROVED"
              class="text-green-500"
            />
            <icon-cross v-else class="text-red-500" />
            <span class="text-gray-600 dark:text-gray-400">In list</span>
          </span>
          <span class="meta-item inline-flex items-center">
            <icon-check v-if="isArtist" class="text-green-500" />
            <icon-cross v-else class="text-red-500" />
            <span class="text-gray-600 dark:text-gray-400">Artist</span>
          </span>
        </div>
      </div>
    </div>

    <div>
      <shared-bsky-description
        v-if="data.description"
        :description="data.description"
      />
    </div>
  </shared-card>
  <shared-card v-else class="bg-red-200 dark:bg-red-700">
    Profile with did {{ did }} was not found.
  </shared-card>
</template>

<style scoped>
.meta .meta-item:not(:last-child)::after {
  content: "·";
  @apply px-1;
  @apply inline-block;
  text-decoration: none;
}

.loading {
  background: linear-gradient(
    120deg,
    transparent 5%,
    rgb(31, 41, 55) 20%,
    transparent 30%
  );
  background-size: 200% 100%;
  background-position-y: bottom;
  animation: 1.25s loading linear infinite;
}

@media (prefers-color-scheme: light) {
  .loading {
    background: linear-gradient(
      120deg,
      transparent 5%,
      rgb(243, 244, 246) 20%,
      transparent 30%
    );
    background-size: 200% 100%;
    background-position-y: bottom;
  }
}

@keyframes loading {
  from {
    background-position-x: 50%;
  }
  to {
    background-position-x: -150%;
  }
}
</style>
