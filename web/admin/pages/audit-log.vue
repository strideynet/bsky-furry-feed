<script setup lang="ts">
const api = await useAPI();

const error = ref<string>();

const { auditEvents } = await api.listAuditEvents({}).catch((err) => {
  error.value = err.rawMessage;

  return {
    auditEvents: [],
  };
});
</script>

<template>
  <div>
    <shared-card v-if="error" variant="error">{{ error }}</shared-card>
    <div v-else>
      <h1 class="text-xl font-bold">Audit log</h1>
      <p class="text-muted">Showing the last 100 audit events.</p>
      <action
        v-for="event in auditEvents"
        :key="event.id"
        :action="event"
        lookup-user
      />
    </div>
  </div>
</template>
