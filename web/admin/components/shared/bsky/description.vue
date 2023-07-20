<script lang="ts" setup>
import { RichText } from "@atproto/api";
import { newAgent } from "~/lib/auth";

const props = defineProps<{ description: string }>();

const segments = ref();

const updateDescription = async () => {
  const descriptionRichText = new RichText(
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
    <shared-bsky-text v-for="segment in segments" :segment="segment" />
    <div v-if="!segments" class="whitespace-pre-line">
      {{ description }}
    </div>
  </div>
</template>
