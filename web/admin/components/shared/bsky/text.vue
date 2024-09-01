<script lang="ts" setup>
import { RichTextSegment } from "@atproto/api";

defineProps<{ segment: RichTextSegment }>();

const hovering = ref(true);

function enter() {
  hovering.value = true;
}

function leave() {
  hovering.value = false;
}
</script>

<template>
  <nuxt-link
    v-if="segment.isLink()"
    class="underline hover:no-underline text-blue-500 break-all"
    :href="segment.link?.uri"
    target="_blank"
  >
    {{ segment.text }}
  </nuxt-link>
  <span v-else-if="segment.isMention()" class="relative">
    <nuxt-link
      class="underline hover:no-underline text-blue-500"
      :href="`https://bsky.app/profile/${segment.mention?.did}`"
      target="_blank"
      @mouseenter="enter"
      @mouseleave="leave"
    >
      {{ segment.text }}
    </nuxt-link>

    <div
      v-if="hovering && segment.mention?.did"
      class="absolute flex left-0 top-0 right-0 w-full"
    >
      <shared-user-tooltip :did="segment.mention.did" />
    </div>
  </span>
  <span v-else class="whitespace-pre-line">{{ segment.text }}</span>
</template>
