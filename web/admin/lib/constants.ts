import { ActorStatus } from "../../proto/bff/v1/types_pb";

export const ACTOR_STATUS_LABELS = {
  [ActorStatus.UNSPECIFIED]: "Unspecified",
  [ActorStatus.PENDING]: "Pending",
  [ActorStatus.APPROVED]: "Approved",
  [ActorStatus.BANNED]: "Banned",
  [ActorStatus.NONE]: "None",
  [ActorStatus.OPTED_OUT]: "Opted out",
  [ActorStatus.REJECTED]: "Rejected",
};
