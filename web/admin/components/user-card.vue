<script lang="ts" setup>
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { Actor, ActorStatus } from "../../proto/bff/v1/types_pb";
import { getProfile } from "~/lib/cached-bsky";
import { newAgent } from "~/lib/auth";
import { ViewImage } from "@atproto/api/dist/client/types/app/bsky/embed/images";
import { PostView } from "@atproto/api/dist/client/types/app/bsky/feed/defs";

const props = defineProps<{
  did: string;
  pending?: number;
  variant: "queue" | "profile";
}>();
const $emit = defineEmits(["next"]);

const api = await useAPI();
const showAvatarModal = ref(false);
const loading = ref(false);
const actor = ref<Actor>();
const data = ref<ProfileViewDetailed>();
const loadProfile = async () => {
  data.value = await getProfile(props.did);
  const response = await api
    .getActor({ did: data.value?.did || props.did })
    .catch(() => ({ actor: undefined }));
  actor.value = response?.actor;

  posts.value = await newAgent()
    .getAuthorFeed({ actor: props.did })
    .then((r) =>
      r.data.feed
        .filter((p) => !p.reply && p.post.author.did === props.did)
        .map((p) => p.post)
    );
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

function addSISuffix(number?: number) {
  number = number || 0;

  const suffixes = ["", "K", "M"];
  const order = Math.floor(Math.log10(number) / 3);

  for (let i = 0; i < order; i++) {
    number = number / 1000;
  }

  return `${Math.round(number * 100) / 100}${suffixes[order] || ""}`;
}

const posts = ref<PostView[]>([]);

await loadProfile();
</script>

<template>
  <template v-if="data">
    <user-queue-banner
      v-if="actor?.status === ActorStatus.PENDING"
      :did="data.did"
      :name="data.displayName || data.handle.replace(/.bsky.social$/, '')"
      :pending="pending"
      @next="next"
      @loading="loading = true"
    />
    <div class="flex max-md:flex-col gap-3" :class="{ loading }">
      <div class="mb-3 md:w-[50%] card-list h-min flex-1">
        <user-actions :did="data.did" :status="actor?.status" @update="next" />
        <shared-card>
          <img
            v-if="data.banner"
            :src="`https://bsky-cdn.codingpa.ws/banner/${did}/450x150`"
            class="w-full object-fit rounded-lg"
            height="101"
            width="304"
            alt=""
          />
        </shared-card>
        <shared-card>
          <div class="flex gap-3 items-center">
            <button
              class="relative flex overflow-hidden rounded-lg"
              @click="showAvatarModal = true"
            >
              <shared-avatar
                :did="data.did"
                :has-avatar="Boolean(data.avatar)"
                not-rounded
                resize="72x72"
                :size="72"
              />
              <span
                class="opacity-0 hover:opacity-100 transition duration-300 w-full h-full absolute flex items-center bg-black bg-opacity-50 text-xs uppercase tracking-tight"
              >
                Click to zoom
              </span>
            </button>
            <core-modal v-if="showAvatarModal" @close="showAvatarModal = false">
              <div class="z-10">
                <shared-avatar
                  class="w-auto h-auto max-h-[80vh] max-w-[80vw]"
                  :did="data.did"
                  :has-avatar="Boolean(data.avatar)"
                  resize="webp"
                  :size="512"
                />
              </div>
            </core-modal>
            <div class="flex flex-col">
              <div v-if="data.displayName" class="text-lg">
                {{ data.displayName }}
              </div>
              <div>
                <nuxt-link
                  class="underline hover:no-underline text-muted"
                  :href="`https://bsky.app/profile/${data.handle}`"
                  target="_blank"
                >
                  @{{ data.handle }}
                </nuxt-link>
              </div>
            </div>
          </div>
        </shared-card>
        <shared-card v-if="actor?.roles" class="flex items-center gap-1 hidden">
          <icon-key class="text-muted" />
          {{ actor.roles.join(", ") }}
          <button
            class="ml-auto text-sm rounded-lg py-0.5 px-1.5 border border-gray-300 dark:border-gray-700 hover:bg-zinc-700"
          >
            Edit
          </button>
        </shared-card>
        <shared-card class="meta">
          <div class="meta-item">
            <user-status-badge class="text-sm" :status="actor?.status" />
          </div>
          <span
            class="ml-auto meta-item"
            :title="`${data?.followersCount || 0} followers`"
          >
            <icon-users class="text-muted" />
            {{ addSISuffix(data?.followersCount) }}
          </span>
          <span class="meta-item" :title="`${data?.followsCount || 0} follows`">
            <icon-user-check class="text-muted" />
            {{ addSISuffix(data?.followsCount) }}
          </span>
          <span class="meta-item" :title="`${data?.postsCount || 0} posts`">
            <icon-square-bubble class="text-muted" :size="18" />
            {{ addSISuffix(data?.postsCount) }}
          </span>
        </shared-card>
        <shared-card v-if="data.description">
          <shared-bsky-description :description="data.description" />
        </shared-card>
      </div>
      <div class="mb-3 md:w-[50%]">
        <shared-card :class="{ loading }" no-padding>
          <div class="px-4 py-3 border-b border-gray-300 dark:border-gray-700">
            <h2>Recent posts</h2>
          </div>
          <div class="overflow-y-auto max-h-[500px]">
            <div
              v-for="post in posts"
              :key="post.uri"
              class="px-4 py-2 border-b border-gray-300 dark:border-gray-700"
            >
              <template v-if="post">
                <div class="meta text-sm text-muted">
                  <span class="meta-item">
                    <shared-date :date="new Date(post.indexedAt)" />
                  </span>
                  <span class="meta-item flex items-center gap-0.5">
                    <icon-heart class="text-muted" />
                    {{ addSISuffix(post.likeCount || 0) }}
                  </span>
                  <span class="meta-item flex items-center gap-0.5">
                    <icon-square-bubble class="text-muted" :size="14" />
                    {{ addSISuffix(post.replyCount || 0) }}
                  </span>
                </div>
                <div class="flex">
                  <span class="flex-1">{{ (post.record as any)?.text }}</span>
                  <span
                    v-if="post.embed && 'images' in post.embed"
                    class="w-[25%] h-100"
                  >
                    <img
                      v-for="img in (post.embed.images as ViewImage[]).slice(0,1)"
                      :key="img.thumb"
                      class="object-cover h-100 rounded-lg"
                      :src="img.thumb"
                      :alt="img.alt"
                    />
                  </span>
                </div>
              </template>
              <div v-else class="text-sm text-muted">
                Error: post not found.
              </div>
            </div>
            <div
              v-if="posts.length === 0"
              class="text-muted px-4 py-2 border-gray-300 dark:border-gray-700"
            >
              No recent posts.
            </div>
          </div>
        </shared-card>
      </div>
    </div>
  </template>
  <div v-else>
    <user-queue-banner
      v-if="actor?.status === ActorStatus.PENDING"
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
.meta {
  @apply flex items-center gap-3;
}
.meta-item {
  @apply inline-flex items-center gap-1;
}

.card-list > :not(:last-of-type) {
  @apply border-b-0;
  @apply rounded-b-none;
}

.card-list > :not(:first-of-type) {
  @apply rounded-t-none;
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
