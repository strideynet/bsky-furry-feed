// @generated by protoc-gen-es v1.2.0 with parameter "target=js+dts"
// @generated from file proto/bff/v1/moderation_service.proto (package bff.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { Any, BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage, Timestamp } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum bff.v1.ActorStatus
 */
export declare enum ActorStatus {
  /**
   * @generated from enum value: ACTOR_STATUS_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: ACTOR_STATUS_PENDING = 1;
   */
  PENDING = 1,

  /**
   * @generated from enum value: ACTOR_STATUS_APPROVED = 2;
   */
  APPROVED = 2,

  /**
   * @generated from enum value: ACTOR_STATUS_BANNED = 3;
   */
  BANNED = 3,

  /**
   * @generated from enum value: ACTOR_STATUS_NONE = 4;
   */
  NONE = 4,
}

/**
 * @generated from enum bff.v1.ApprovalQueueAction
 */
export declare enum ApprovalQueueAction {
  /**
   * @generated from enum value: APPROVAL_QUEUE_ACTION_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: APPROVAL_QUEUE_ACTION_APPROVE = 1;
   */
  APPROVE = 1,

  /**
   * @generated from enum value: APPROVAL_QUEUE_ACTION_REJECT = 2;
   */
  REJECT = 2,
}

/**
 * @generated from message bff.v1.Actor
 */
export declare class Actor extends Message<Actor> {
  /**
   * did is the decentralized identity of the actor. This is also the UID used
   * for fetching and mutating actors.
   *
   * @generated from field: string did = 1;
   */
  did: string;

  /**
   * is_hidden is a deprecated flag that used to hide accounts. This no longer
   * has any effect.
   * Deprecated: Use status.
   *
   * @generated from field: bool is_hidden = 2;
   */
  isHidden: boolean;

  /**
   * is_artist is a flag indicating this account is primarily an artist. It
   * does not currently control any feed placement.
   *
   * @generated from field: bool is_artist = 3;
   */
  isArtist: boolean;

  /**
   * comment is a short string that is applied to an account when it is added
   * to the system. This will eventually be replaced by a more powerful system.
   *
   * @generated from field: string comment = 4;
   */
  comment: string;

  /**
   * status indicates the actor's current status.
   *
   * @generated from field: bff.v1.ActorStatus status = 5;
   */
  status: ActorStatus;

  /**
   * created_at indicates the time that the actor was first added to the bff
   * system - this does not necessarily indicate when they joined bluesky.
   *
   * @generated from field: google.protobuf.Timestamp created_at = 6;
   */
  createdAt?: Timestamp;

  constructor(data?: PartialMessage<Actor>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.Actor";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Actor;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Actor;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Actor;

  static equals(a: Actor | PlainMessage<Actor> | undefined, b: Actor | PlainMessage<Actor> | undefined): boolean;
}

/**
 * @generated from message bff.v1.Post
 */
export declare class Post extends Message<Post> {
  /**
   * @generated from field: string uri = 1;
   */
  uri: string;

  /**
   * @generated from field: google.protobuf.Timestamp created_at = 2;
   */
  createdAt?: Timestamp;

  /**
   * @generated from field: google.protobuf.Timestamp indexed_at = 3;
   */
  indexedAt?: Timestamp;

  constructor(data?: PartialMessage<Post>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.Post";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Post;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Post;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Post;

  static equals(a: Post | PlainMessage<Post> | undefined, b: Post | PlainMessage<Post> | undefined): boolean;
}

/**
 * @generated from message bff.v1.GetActorRequest
 */
export declare class GetActorRequest extends Message<GetActorRequest> {
  /**
   * @generated from field: string did = 1;
   */
  did: string;

  constructor(data?: PartialMessage<GetActorRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.GetActorRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetActorRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetActorRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetActorRequest;

  static equals(a: GetActorRequest | PlainMessage<GetActorRequest> | undefined, b: GetActorRequest | PlainMessage<GetActorRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.GetActorResponse
 */
export declare class GetActorResponse extends Message<GetActorResponse> {
  /**
   * @generated from field: bff.v1.Actor actor = 1;
   */
  actor?: Actor;

  constructor(data?: PartialMessage<GetActorResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.GetActorResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetActorResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetActorResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetActorResponse;

  static equals(a: GetActorResponse | PlainMessage<GetActorResponse> | undefined, b: GetActorResponse | PlainMessage<GetActorResponse> | undefined): boolean;
}

/**
 * @generated from message bff.v1.ListActorsRequest
 */
export declare class ListActorsRequest extends Message<ListActorsRequest> {
  /**
   * @generated from field: string cursor = 1;
   */
  cursor: string;

  /**
   * @generated from field: int32 limit = 2;
   */
  limit: number;

  /**
   * @generated from field: bff.v1.ActorStatus filter_status = 3;
   */
  filterStatus: ActorStatus;

  constructor(data?: PartialMessage<ListActorsRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.ListActorsRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListActorsRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListActorsRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListActorsRequest;

  static equals(a: ListActorsRequest | PlainMessage<ListActorsRequest> | undefined, b: ListActorsRequest | PlainMessage<ListActorsRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.ListActorsResponse
 */
export declare class ListActorsResponse extends Message<ListActorsResponse> {
  /**
   * @generated from field: repeated bff.v1.Actor actors = 1;
   */
  actors: Actor[];

  /**
   * @generated from field: string cursor = 2;
   */
  cursor: string;

  constructor(data?: PartialMessage<ListActorsResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.ListActorsResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListActorsResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListActorsResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListActorsResponse;

  static equals(a: ListActorsResponse | PlainMessage<ListActorsResponse> | undefined, b: ListActorsResponse | PlainMessage<ListActorsResponse> | undefined): boolean;
}

/**
 * @generated from message bff.v1.PingRequest
 */
export declare class PingRequest extends Message<PingRequest> {
  constructor(data?: PartialMessage<PingRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.PingRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PingRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PingRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PingRequest;

  static equals(a: PingRequest | PlainMessage<PingRequest> | undefined, b: PingRequest | PlainMessage<PingRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.PingResponse
 */
export declare class PingResponse extends Message<PingResponse> {
  constructor(data?: PartialMessage<PingResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.PingResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PingResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PingResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PingResponse;

  static equals(a: PingResponse | PlainMessage<PingResponse> | undefined, b: PingResponse | PlainMessage<PingResponse> | undefined): boolean;
}

/**
 * @generated from message bff.v1.GetApprovalQueueRequest
 */
export declare class GetApprovalQueueRequest extends Message<GetApprovalQueueRequest> {
  constructor(data?: PartialMessage<GetApprovalQueueRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.GetApprovalQueueRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetApprovalQueueRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetApprovalQueueRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetApprovalQueueRequest;

  static equals(a: GetApprovalQueueRequest | PlainMessage<GetApprovalQueueRequest> | undefined, b: GetApprovalQueueRequest | PlainMessage<GetApprovalQueueRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.GetApprovalQueueResponse
 */
export declare class GetApprovalQueueResponse extends Message<GetApprovalQueueResponse> {
  /**
   * queue_entry is the actor that needs to be processed by a mod. process the
   * queue entry using the ProcessApprovalQueue RPC.
   *
   * @generated from field: bff.v1.Actor queue_entry = 1;
   */
  queueEntry?: Actor;

  /**
   * queue_entries_remaining indicates how many queue entries are left including
   * the one returned in this response.
   *
   * @generated from field: int32 queue_entries_remaining = 2;
   */
  queueEntriesRemaining: number;

  constructor(data?: PartialMessage<GetApprovalQueueResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.GetApprovalQueueResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetApprovalQueueResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetApprovalQueueResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetApprovalQueueResponse;

  static equals(a: GetApprovalQueueResponse | PlainMessage<GetApprovalQueueResponse> | undefined, b: GetApprovalQueueResponse | PlainMessage<GetApprovalQueueResponse> | undefined): boolean;
}

/**
 * @generated from message bff.v1.ProcessApprovalQueueRequest
 */
export declare class ProcessApprovalQueueRequest extends Message<ProcessApprovalQueueRequest> {
  /**
   * @generated from field: string did = 1;
   */
  did: string;

  /**
   * @generated from field: bff.v1.ApprovalQueueAction action = 2;
   */
  action: ApprovalQueueAction;

  /**
   * @generated from field: bool is_artist = 3;
   */
  isArtist: boolean;

  constructor(data?: PartialMessage<ProcessApprovalQueueRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.ProcessApprovalQueueRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProcessApprovalQueueRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProcessApprovalQueueRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProcessApprovalQueueRequest;

  static equals(a: ProcessApprovalQueueRequest | PlainMessage<ProcessApprovalQueueRequest> | undefined, b: ProcessApprovalQueueRequest | PlainMessage<ProcessApprovalQueueRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.ProcessApprovalQueueResponse
 */
export declare class ProcessApprovalQueueResponse extends Message<ProcessApprovalQueueResponse> {
  constructor(data?: PartialMessage<ProcessApprovalQueueResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.ProcessApprovalQueueResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProcessApprovalQueueResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProcessApprovalQueueResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProcessApprovalQueueResponse;

  static equals(a: ProcessApprovalQueueResponse | PlainMessage<ProcessApprovalQueueResponse> | undefined, b: ProcessApprovalQueueResponse | PlainMessage<ProcessApprovalQueueResponse> | undefined): boolean;
}

/**
 * @generated from message bff.v1.AuditEntry
 */
export declare class AuditEntry extends Message<AuditEntry> {
  /**
   * @generated from field: string id = 1;
   */
  id: string;

  /**
   * @generated from field: google.protobuf.Timestamp created_at = 2;
   */
  createdAt?: Timestamp;

  /**
   * @generated from field: string creator_did = 3;
   */
  creatorDid: string;

  /**
   * @generated from field: string subject_actor_did = 4;
   */
  subjectActorDid: string;

  /**
   * subject_uri allows an audit entry to be linked to a specific bluesky
   * entity. This is an optional field and can be empty if targetting just
   * an actor.
   *
   * @generated from field: optional string subject_uri = 5;
   */
  subjectUri?: string;

  /**
   * comment is a comment left by the creator of the audit event to explain
   * the action.
   *
   * @generated from field: string comment = 6;
   */
  comment: string;

  /**
   * @generated from field: google.protobuf.Any payload = 7;
   */
  payload?: Any;

  constructor(data?: PartialMessage<AuditEntry>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.AuditEntry";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AuditEntry;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AuditEntry;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AuditEntry;

  static equals(a: AuditEntry | PlainMessage<AuditEntry> | undefined, b: AuditEntry | PlainMessage<AuditEntry> | undefined): boolean;
}

/**
 * @generated from message bff.v1.QueueApprovalAuditPayload
 */
export declare class QueueApprovalAuditPayload extends Message<QueueApprovalAuditPayload> {
  /**
   * @generated from field: bff.v1.ApprovalQueueAction action = 1;
   */
  action: ApprovalQueueAction;

  /**
   * @generated from field: bool is_artist = 2;
   */
  isArtist: boolean;

  constructor(data?: PartialMessage<QueueApprovalAuditPayload>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.QueueApprovalAuditPayload";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): QueueApprovalAuditPayload;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): QueueApprovalAuditPayload;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): QueueApprovalAuditPayload;

  static equals(a: QueueApprovalAuditPayload | PlainMessage<QueueApprovalAuditPayload> | undefined, b: QueueApprovalAuditPayload | PlainMessage<QueueApprovalAuditPayload> | undefined): boolean;
}

