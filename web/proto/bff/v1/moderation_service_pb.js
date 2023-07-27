// @generated by protoc-gen-es v1.2.0 with parameter "target=js+dts"
// @generated from file proto/bff/v1/moderation_service.proto (package bff.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { Any, proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from enum bff.v1.ActorStatus
 */
export const ActorStatus = proto3.makeEnum(
  "bff.v1.ActorStatus",
  [
    {no: 0, name: "ACTOR_STATUS_UNSPECIFIED", localName: "UNSPECIFIED"},
    {no: 1, name: "ACTOR_STATUS_PENDING", localName: "PENDING"},
    {no: 2, name: "ACTOR_STATUS_APPROVED", localName: "APPROVED"},
    {no: 3, name: "ACTOR_STATUS_BANNED", localName: "BANNED"},
    {no: 4, name: "ACTOR_STATUS_NONE", localName: "NONE"},
  ],
);

/**
 * @generated from enum bff.v1.ApprovalQueueAction
 */
export const ApprovalQueueAction = proto3.makeEnum(
  "bff.v1.ApprovalQueueAction",
  [
    {no: 0, name: "APPROVAL_QUEUE_ACTION_UNSPECIFIED", localName: "UNSPECIFIED"},
    {no: 1, name: "APPROVAL_QUEUE_ACTION_APPROVE", localName: "APPROVE"},
    {no: 2, name: "APPROVAL_QUEUE_ACTION_REJECT", localName: "REJECT"},
  ],
);

/**
 * @generated from message bff.v1.Actor
 */
export const Actor = proto3.makeMessageType(
  "bff.v1.Actor",
  () => [
    { no: 1, name: "did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "is_hidden", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 3, name: "is_artist", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 4, name: "comment", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "status", kind: "enum", T: proto3.getEnumType(ActorStatus) },
    { no: 6, name: "created_at", kind: "message", T: Timestamp },
  ],
);

/**
 * @generated from message bff.v1.Post
 */
export const Post = proto3.makeMessageType(
  "bff.v1.Post",
  () => [
    { no: 1, name: "uri", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "created_at", kind: "message", T: Timestamp },
    { no: 3, name: "indexed_at", kind: "message", T: Timestamp },
  ],
);

/**
 * @generated from message bff.v1.GetActorRequest
 */
export const GetActorRequest = proto3.makeMessageType(
  "bff.v1.GetActorRequest",
  () => [
    { no: 1, name: "did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message bff.v1.GetActorResponse
 */
export const GetActorResponse = proto3.makeMessageType(
  "bff.v1.GetActorResponse",
  () => [
    { no: 1, name: "actor", kind: "message", T: Actor },
  ],
);

/**
 * @generated from message bff.v1.ListActorsRequest
 */
export const ListActorsRequest = proto3.makeMessageType(
  "bff.v1.ListActorsRequest",
  () => [
    { no: 1, name: "cursor", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "limit", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 3, name: "filter_status", kind: "enum", T: proto3.getEnumType(ActorStatus) },
  ],
);

/**
 * @generated from message bff.v1.ListActorsResponse
 */
export const ListActorsResponse = proto3.makeMessageType(
  "bff.v1.ListActorsResponse",
  () => [
    { no: 1, name: "actors", kind: "message", T: Actor, repeated: true },
    { no: 2, name: "cursor", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message bff.v1.PingRequest
 */
export const PingRequest = proto3.makeMessageType(
  "bff.v1.PingRequest",
  [],
);

/**
 * @generated from message bff.v1.PingResponse
 */
export const PingResponse = proto3.makeMessageType(
  "bff.v1.PingResponse",
  [],
);

/**
 * @generated from message bff.v1.GetApprovalQueueRequest
 */
export const GetApprovalQueueRequest = proto3.makeMessageType(
  "bff.v1.GetApprovalQueueRequest",
  [],
);

/**
 * @generated from message bff.v1.GetApprovalQueueResponse
 */
export const GetApprovalQueueResponse = proto3.makeMessageType(
  "bff.v1.GetApprovalQueueResponse",
  () => [
    { no: 1, name: "queue_entry", kind: "message", T: Actor },
    { no: 2, name: "queue_entries_remaining", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ],
);

/**
 * @generated from message bff.v1.ProcessApprovalQueueRequest
 */
export const ProcessApprovalQueueRequest = proto3.makeMessageType(
  "bff.v1.ProcessApprovalQueueRequest",
  () => [
    { no: 1, name: "did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "action", kind: "enum", T: proto3.getEnumType(ApprovalQueueAction) },
    { no: 3, name: "is_artist", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ],
);

/**
 * @generated from message bff.v1.ProcessApprovalQueueResponse
 */
export const ProcessApprovalQueueResponse = proto3.makeMessageType(
  "bff.v1.ProcessApprovalQueueResponse",
  [],
);

/**
 * @generated from message bff.v1.ListAuditEventsRequest
 */
export const ListAuditEventsRequest = proto3.makeMessageType(
  "bff.v1.ListAuditEventsRequest",
  () => [
    { no: 1, name: "actor_did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "subject_did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "subject_record_uri", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message bff.v1.ListAuditEventsResponse
 */
export const ListAuditEventsResponse = proto3.makeMessageType(
  "bff.v1.ListAuditEventsResponse",
  () => [
    { no: 1, name: "audit_events", kind: "message", T: AuditEvent, repeated: true },
  ],
);

/**
 * @generated from message bff.v1.CreateCommentAuditEventRequest
 */
export const CreateCommentAuditEventRequest = proto3.makeMessageType(
  "bff.v1.CreateCommentAuditEventRequest",
  () => [
    { no: 1, name: "subject_did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "subject_record_uri", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "comment", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message bff.v1.CreateCommentAuditEventResponse
 */
export const CreateCommentAuditEventResponse = proto3.makeMessageType(
  "bff.v1.CreateCommentAuditEventResponse",
  () => [
    { no: 1, name: "audit_event", kind: "message", T: AuditEvent },
  ],
);

/**
 * @generated from message bff.v1.AuditEvent
 */
export const AuditEvent = proto3.makeMessageType(
  "bff.v1.AuditEvent",
  () => [
    { no: 1, name: "id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "created_at", kind: "message", T: Timestamp },
    { no: 3, name: "actor_did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "subject_did", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "subject_record_uri", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "payload", kind: "message", T: Any },
  ],
);

/**
 * CommentAuditPayload is the payload for the `comment`audit event. This is
 * empty, as the comment is actually held within `AuditEvent`
 *
 * @generated from message bff.v1.CommentAuditPayload
 */
export const CommentAuditPayload = proto3.makeMessageType(
  "bff.v1.CommentAuditPayload",
  () => [
    { no: 1, name: "comment", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ProcessApprovalQueueAuditPayload is the payload for the
 * `process_approval_queue` audit event.
 *
 * @generated from message bff.v1.ProcessApprovalQueueAuditPayload
 */
export const ProcessApprovalQueueAuditPayload = proto3.makeMessageType(
  "bff.v1.ProcessApprovalQueueAuditPayload",
  () => [
    { no: 1, name: "action", kind: "enum", T: proto3.getEnumType(ApprovalQueueAction) },
  ],
);

