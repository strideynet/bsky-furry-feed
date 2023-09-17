// @generated by protoc-gen-es v1.2.0 with parameter "target=js+dts"
// @generated from file bff/v1/types.proto (package bff.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage, Timestamp } from "@bufbuild/protobuf";
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

  /**
   * roles is an array of roles this actor holds in relation to actions on the
   * moderation API.
   *
   * @generated from field: repeated string roles = 7;
   */
  roles: string[];

  /**
   * in_queue_after is the time after which an actor with the PENDING status
   * is available to be processed in the queue
   *
   * @generated from field: google.protobuf.Timestamp in_queue_after = 8;
   */
  inQueueAfter?: Timestamp;

  constructor(data?: PartialMessage<Actor>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.Actor";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Actor;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Actor;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Actor;

  static equals(a: Actor | PlainMessage<Actor> | undefined, b: Actor | PlainMessage<Actor> | undefined): boolean;
}

