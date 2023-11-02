<script setup lang="ts">
import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { Actor } from "../../../proto/bff/v1/types_pb";

defineProps<{
  users: (Actor & ProfileViewDetailed)[];
}>();
</script>

<template>
  <div>
    <held-back-user
      v-for="(user, index) in [...users].sort(
        (a, b) =>
          (a.heldUntil?.toDate().getTime() || 0) -
          (b.heldUntil?.toDate().getTime() || 0)
      )"
      :key="user.did"
      :class="{
        'rounded-t': index === 0,
        'rounded-b': index === users.length - 1,
        'border-b-0': index > 0 && index < users.length - 1,
      }"
      :user="user"
    />
  </div>
</template>
