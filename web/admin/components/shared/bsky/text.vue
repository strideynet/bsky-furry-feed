<script lang="ts" setup>
import { RichTextSegment } from "@atproto/api";

defineProps<{ segment: RichTextSegment }>();
</script>

<template>
  <nuxt-link
    v-if="segment.isLink()"
    class="underline hover:no-underline text-blue-500"
    :href="segment.link?.uri"
    target="_blank"
  >
    {{ segment.text }}
  </nuxt-link>
  <nuxt-link
    v-else-if="segment.isMention()"
    class="underline hover:no-underline text-blue-500"
    :href="`https://bsky.app/profile/${segment.mention?.did}`"
    target="_blank"
  >
    {{ segment.text }}
  </nuxt-link>
  <span v-else class="whitespace-pre-line">{{ segment.text }}</span>
</template>
