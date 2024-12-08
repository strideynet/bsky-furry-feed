<script lang="ts" setup>
import { getProfile } from "~/lib/cached-bsky";
import { isProbablyFurry } from "~/lib/furry-detector";
import { Actor, ActorStatus } from "../../proto/bff/v1/types_pb";
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";

const api = await useAPI();

const actors = ref<Actor[]>([]);
const heldBack = ref<Actor[]>([]);

const actorProfiles = ref<ProfileViewDetailed[]>([]);
const randomActor = ref<Actor>();
const currentQueue = ref<keyof typeof queues["value"]>("All");

const error = ref<string>();

const actorProfilesMap = computed(() => {
  const map = new Map<string, ProfileViewDetailed>();

  for (const profile of actorProfiles.value) {
    map.set(profile.did, profile);
  }

  return map;
});

const avatarStockMap = ref(new Map<string, boolean>());

watch(actorProfiles, async () => {
  for (const profile of actorProfiles.value) {
    const { isStockAvatar } = await fetch(
      `https://bsky-cdn.codingpa.ws/avatar/${profile.did}/info.json`
    ).then((r) => r.json());
    avatarStockMap.value.set(profile.did, isStockAvatar);
  }
});

const queues = computed(() => ({
  All: actors.value,
  "Probably furry": actors.value.filter((actor) => {
    const profile = didToProfile(actor.did);

    return isProbablyFurry(profile) && !avatarStockMap.value.get(actor.did);
  }),
  "Empty profiles": actors.value.filter((actor) => {
    const profile = didToProfile(actor.did);
    if (!profile) return false;

    return (
      (!profile.displayName && !profile.description && !profile.postsCount) ||
      avatarStockMap.value.get(actor.did)
    );
  }),
  "Held back": heldBack.value,
}));

const nextActor = async () => {
  error.value = "";

  const queue = await api
    .listActors({ filterStatus: ActorStatus.PENDING })
    .catch((err) => {
      error.value = err.rawMessage;

      return {
        actors: [],
      };
    });

  function isHeldBack(a: Actor) {
    return a.heldUntil && a.heldUntil.toDate() > new Date();
  }

  actors.value = queue.actors.filter((a) => !isHeldBack(a));
  heldBack.value = queue.actors.filter(isHeldBack);
  actorProfiles.value = await Promise.all(
    actors.value.map((a) => a.did).map(getProfile)
  ).then((p) => p.filter(Boolean));
  selectRandomActor();
};

await nextActor();

function didToProfile(did: string): ProfileViewDetailed | undefined {
  return actorProfilesMap.value.get(did);
}

function selectRandomActor() {
  if (!queues.value) return;
  const queue = queues.value[currentQueue.value];

  randomActor.value = queue[Math.floor(Math.random() * queue.length)] as Actor;
}
</script>

<template>
  <div>
    <div class="flex gap-1 mb-3">
      <button
        v-for="(list, label) in queues"
        :key="label"
        class="border hover:bg-slate-800 border-gray-300 dark:border-gray-700 rounded-lg px-2 py-0.5 max-md:text-sm"
        :class="{
          'bg-slate-800': currentQueue === label,
          'ml-auto text-muted': label === 'Held back',
        }"
        @click="
          () => {
            currentQueue = label;
            selectRandomActor();
          }
        "
      >
        {{ label }} ({{ list.length }})
      </button>
    </div>

    <held-back-list
      v-if="currentQueue === 'Held back'"
      :users="
        heldBack.map((actor) => ({ ...actor, ...didToProfile(actor.did) } as Actor & ProfileViewDetailed))
      "
    />
    <div v-else>
      <shared-card v-if="error" variant="error">{{ error }}</shared-card>
      <user-profile
        v-else-if="randomActor"
        :did="randomActor.did"
        :pending="queues[currentQueue].length"
        variant="queue"
        @next="nextActor"
      />
      <shared-card v-else>
        No user is in {{ currentQueue === "All" ? "the" : "this" }} queue.
      </shared-card>
    </div>
  </div>
</template>
