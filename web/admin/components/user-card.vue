<script lang="ts" setup>
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { ActorStatus } from "../../proto/bff/v1/types_pb";
import { getProfile } from "~/lib/cached-bsky";

const props = defineProps<{
  did: string;
  pending?: number;
  variant: "queue" | "profile";
}>();
const $emit = defineEmits(["next"]);

const api = await useAPI();
const isArtist = ref(false);
const loading = ref(false);
const status = ref<ActorStatus>();
const data = ref<ProfileViewDetailed>();
const loadProfile = async () => {
  data.value = await getProfile(props.did);
  const { actor } = await api
    .getActor({ did: data.value?.did || props.did })
    .catch(() => ({ actor: undefined }));
  isArtist.value = Boolean(actor?.isArtist);
  status.value = actor?.status;
};

async function next() {
  if (props.variant === "profile") {
    await loadProfile();
  }
  loading.value = false;
  $emit("next");
}

watch(
  () => props.did,
  () => loadProfile()
);

await loadProfile();
</script>

<template>
  <shared-card v-if="data" :class="{ loading }">
    <user-queue-actions
      v-if="status === ActorStatus.PENDING"
      :did="data.did"
      :name="data.displayName || data.handle.replace(/.bsky.social$/, '')"
      :pending="pending"
      @next="next"
      @loading="loading = true"
    />
    <user-actions
      v-if="variant === 'profile'"
      :did="data.did"
      :status="status"
      @update="next"
    />

    <div class="flex gap-3 items-center mb-5">
      <shared-avatar :url="data.avatar" :size="72" />
      <div class="flex flex-col">
        <div class="text-lg">{{ data.displayName || data.handle }}</div>
        <div class="meta">
          <span class="meta-item">
            <nuxt-link
              class="underline hover:no-underline text-muted"
              :href="`https://bsky.app/profile/${data.handle}`"
              target="_blank"
              >@{{ data.handle }}</nuxt-link
            >
          </span>
          <span class="meta-item">
            {{ data.followersCount }}
            <span class="text-muted">followers</span>
          </span>
          <span class="meta-item">
            {{ data.followsCount }}
            <span class="text-muted">follows</span>
          </span>
          <span class="meta-item">
            {{ data.postsCount }}
            <span class="text-muted">posts</span>
          </span>
        </div>
        <div v-if="variant === 'profile'" class="meta">
          <span class="meta-item inline-flex items-center">
            <icon-check v-if="isArtist" class="text-green-500" />
            <icon-cross v-else class="text-red-500" />
            <span class="text-muted">Artist</span>
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
  <div v-else>
    <user-queue-actions
      v-if="status === ActorStatus.PENDING"
      :did="props.did"
      :name="props.did"
      :pending="pending"
      reject-only
      @next="next"
      @loading="loading = true"
    />
    <shared-card class="bg-red-200 dark:bg-red-700">
      Profile with did {{ did }} was not found.
    </shared-card>
  </div>
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
