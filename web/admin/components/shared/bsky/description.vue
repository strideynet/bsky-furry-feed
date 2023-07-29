<script lang="ts" setup>
import * as atproto from "@atproto/api";
import { newAgent } from "~/lib/auth";

const props = defineProps<{ description: string }>();

const segments = ref();

const updateDescription = async () => {
  const descriptionRichText = new atproto.RichText(
    { text: props.description },
    { cleanNewlines: true }
  );
  await descriptionRichText.detectFacets(newAgent());
  segments.value = [...descriptionRichText.segments()];
};

onMounted(updateDescription);
watch(() => props.description, updateDescription);
</script>
<template>
  <div>
    <shared-bsky-text
      v-for="(segment, index) in segments"
      :key="index"
      :segment="segment"
    />
    <div v-if="!segments" class="whitespace-pre-line">
      {{ description }}
    </div>
  </div>
</template>
