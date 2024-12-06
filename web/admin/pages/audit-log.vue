<script setup lang="ts">
import Multiselect from "vue-multiselect";
import "vue-multiselect/dist/vue-multiselect.min.css";
import {
  AuditEvent,
  AuditEventType,
} from "../../proto/bff/v1/moderation_service_pb";

const api = await useAPI();

const types = ref<Array<{ key: AuditEventType; value: string }>>([]);

const error = ref<string>();

const auditEvents: Ref<AuditEvent[]> = ref([]);

async function fetchAuditEvents() {
  const resp = await api
    .listAuditEvents({
      filterTypes: types.value.map((a) => a.key),
    })
    .catch((err) => {
      error.value = err.rawMessage;

      return {
        auditEvents: [],
      };
    });

  auditEvents.value = resp.auditEvents;
}

await fetchAuditEvents();

watch(types, () => fetchAuditEvents());
</script>

<template>
  <div>
    <shared-card v-if="error" variant="error">{{ error }}</shared-card>
    <div v-else>
      <h1 class="text-xl font-bold">Audit log</h1>
      <p class="text-muted mb-1">Showing the last 100 audit events.</p>
      <div class="mb-2">
        <div class="flex flex-col gap-0.5 max-w-[400px]">
          <label for="filter-types" class="font-bold">Types</label>
          <Multiselect
            v-model="types"
            name="filter-types"
            multiple
            track-by="key"
            :options="Object.entries(AuditEventType).filter(t => isNaN(t[0] as any)).map(([label, key]) => ({key, label}))"
            :custom-label="
              (option: any) => option.label.toLowerCase().replace(/^\w/, (a:string) => a.toUpperCase()).replace(/_/g, ' ')
            "
            placeholder="Select type"
          />
        </div>
      </div>
      <action
        v-for="event in auditEvents"
        :key="event.id"
        :action="event"
        lookup-user
      />
    </div>
  </div>
</template>

<style>
.multiselect__tag,
.multiselect__option--highlight,
.multiselect__option--highlight::after {
  @apply bg-blue-500;
}
</style>
