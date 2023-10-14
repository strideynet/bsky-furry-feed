<script setup lang="ts">
import { showQueueActionConfirmation } from "~/lib/settings";
import { logout } from "~/lib/auth";
</script>

<template>
  <div>
    <h1 class="text-xl font-bold">My settings</h1>
    <p class="text-muted mb-4">Settings aren’t synced across your devices.</p>

    <div class="flex gap-2 mb-4">
      <input
        id="show-queue-action-confirmation"
        v-model="showQueueActionConfirmation"
        type="checkbox"
      />
      <label for="show-queue-action-confirmation"
        >Show confirmation popup when approving and rejecting in the
        queue.</label
      >
    </div>

    <button class="text-white bg-gray-700 px-2 py-1 rounded-lg" @click="logout">
      Logout
    </button>
  </div>
</template>
