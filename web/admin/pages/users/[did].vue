<script setup lang="ts">
const did = ref(String(useRoute().params.did));
watch(
  () => String(useRoute().params.did),
  (newDid) => {
    if (newDid) {
      did.value = newDid;
    }
  }
);
</script>

<template>
  <user-profile :did="did" variant="profile" />
</template>
