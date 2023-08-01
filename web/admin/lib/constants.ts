import { ActorStatus } from "../../proto/bff/v1/moderation_service_pb";

export const ACTOR_STATUS_LABELS = {
  [ActorStatus.UNSPECIFIED]: "Unspecified",
  [ActorStatus.PENDING]: "Pending",
  [ActorStatus.APPROVED]: "Approved",
  [ActorStatus.BANNED]: "Banned",
  [ActorStatus.NONE]: "None",
};
