// @generated by protoc-gen-es v1.2.0 with parameter "target=js+dts"
// @generated from file proto/bff/moderation/v1/moderation_service.proto (package bff.moderation.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum bff.moderation.v1.ApprovalQueueAction
 */
export const ApprovalQueueAction = proto3.makeEnum(
  "bff.moderation.v1.ApprovalQueueAction",
  [
    {no: 0, name: "APPROVAL_QUEUE_ACTION_UNSPECIFIED", localName: "UNSPECIFIED"},
    {no: 1, name: "APPROVAL_QUEUE_ACTION_APPROVE", localName: "APPROVE"},
    {no: 2, name: "APPROVAL_QUEUE_ACTION_REJECT", localName: "REJECT"},
  ],
);

/**
 * @generated from message bff.moderation.v1.PingRequest
 */
export const PingRequest = proto3.makeMessageType(
  "bff.moderation.v1.PingRequest",
  [],
);

/**
 * @generated from message bff.moderation.v1.PingResponse
 */
export const PingResponse = proto3.makeMessageType(
  "bff.moderation.v1.PingResponse",
  [],
);

/**
 * @generated from message bff.moderation.v1.CandidateActor
 */
export const CandidateActor = proto3.makeMessageType(
  "bff.moderation.v1.CandidateActor",
  () => [
    { no: 1, name: "did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "is_hidden", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 3, name: "is_artist", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 4, name: "comment", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message bff.moderation.v1.GetApprovalQueueRequest
 */
export const GetApprovalQueueRequest = proto3.makeMessageType(
  "bff.moderation.v1.GetApprovalQueueRequest",
  [],
);

/**
 * @generated from message bff.moderation.v1.GetApprovalQueueResponse
 */
export const GetApprovalQueueResponse = proto3.makeMessageType(
  "bff.moderation.v1.GetApprovalQueueResponse",
  () => [
    { no: 1, name: "queue_entry", kind: "message", T: CandidateActor },
    { no: 2, name: "queue_entries_remaining", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ],
);

/**
 * @generated from message bff.moderation.v1.ProcessApprovalQueueRequest
 */
export const ProcessApprovalQueueRequest = proto3.makeMessageType(
  "bff.moderation.v1.ProcessApprovalQueueRequest",
  () => [
    { no: 1, name: "did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "action", kind: "enum", T: proto3.getEnumType(ApprovalQueueAction) },
    { no: 3, name: "is_artist", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ],
);

/**
 * @generated from message bff.moderation.v1.ProcessApprovalQueueResponse
 */
export const ProcessApprovalQueueResponse = proto3.makeMessageType(
  "bff.moderation.v1.ProcessApprovalQueueResponse",
  [],
);

