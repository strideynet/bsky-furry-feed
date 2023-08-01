<script setup lang="ts">
const api = await useAPI();

const { auditEvents } = await api.listAuditEvents({});
</script>

<template>
  <div>
    <h1 class="text-xl font-bold">Audit log</h1>
    <p class="text-gray-600 dark:text-gray-400">
      Showing the last 100 audit events.
    </p>
    <action
      v-for="event in auditEvents"
      :key="event.id"
      :action="event"
      lookup-user
    />
  </div>
</template>
