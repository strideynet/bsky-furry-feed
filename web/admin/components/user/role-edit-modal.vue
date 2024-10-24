<script setup lang="ts">
import { Actor } from "../../../proto/bff/v1/types_pb";

const $emit = defineEmits<{
  (event: "cancel"): void;
  (event: "update", roles: string[]): void;
}>();

const api = await useAPI();
const allRoles = await api.listRoles({});

const props = defineProps<{ actor: Actor }>();

const loading = ref(false);
const error = ref("");
const roles = ref([...props.actor.roles]);

async function updateRole() {
  error.value = "";
  loading.value = true;
  try {
    await api.assignRoles({ actorDid: props.actor.did, roles: roles.value });
  } catch (err: any) {
    loading.value = false;
    error.value = "message" in err ? err.message : String(err);
    return;
  }
  $emit("update", roles.value);
}
</script>

<template>
  <core-modal @close="$emit('cancel')">
    <div class="z-10"></div>
    <shared-card class="dark:text-white z-10 bg-white dark:bg-gray-900">
      <h2 class="text-lg font-bold">Update role</h2>
      <p class="text-muted mb-3">
        Please select all applicable reasons for the rejection.
      </p>

      <shared-card v-if="error" class="mb-3" variant="error">
        {{ error }}
      </shared-card>

      <ul class="mb-3">
        <li
          v-for="key in Object.keys(allRoles.roles)"
          :key="key"
          class="flex items-center gap-2 border border-gray-300 dark:border-gray-700 rounded-lg px-3 mb-1.5"
        >
          <input
            :id="`role-${key}`"
            type="checkbox"
            :disabled="loading"
            :checked="roles.includes(key)"
            @input="
              () =>
                roles.includes(key)
                  ? (roles = roles.filter((r) => r !== key))
                  : roles.push(key)
            "
          />
          <label
            class="w-full h-full cursor-pointer py-1.5 capitalize"
            :for="`role-${key}`"
          >
            {{ key }}
          </label>
        </li>
      </ul>

      <div class="flex justify-between">
        <button
          class="py-1 whitespace-nowrap px-2 text-white bg-gray-500 dark:bg-gray-600 hover:bg-gray-600 dark:hover:bg-gray-700 disabled:bg-gray-400 disabled:dark:bg-gray-500 rounded-lg disabled:cursor-not-allowed"
          @click="$emit('cancel')"
        >
          Cancel
        </button>

        <button
          class="py-1 px-2 mr-1 bg-blue-400 dark:bg-blue-600 hover:bg-blue-500 dark:hover:bg-blue-700 disabled:bg-blue-300 disabled:dark:bg-blue-500 rounded-lg disabled:cursor-not-allowed"
          :disabled="loading"
          @click="updateRole"
        >
          Update
        </button>
      </div>
    </shared-card>
  </core-modal>
</template>
