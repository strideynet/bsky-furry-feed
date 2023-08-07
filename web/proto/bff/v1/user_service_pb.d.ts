// @generated by protoc-gen-es v1.2.0 with parameter "target=js+dts"
// @generated from file bff/v1/user_service.proto (package bff.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * @generated from message bff.v1.GetMeRequest
 */
export declare class GetMeRequest extends Message<GetMeRequest> {
  constructor(data?: PartialMessage<GetMeRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.GetMeRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetMeRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetMeRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetMeRequest;

  static equals(a: GetMeRequest | PlainMessage<GetMeRequest> | undefined, b: GetMeRequest | PlainMessage<GetMeRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.GetMeResponse
 */
export declare class GetMeResponse extends Message<GetMeResponse> {
  constructor(data?: PartialMessage<GetMeResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.GetMeResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetMeResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetMeResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetMeResponse;

  static equals(a: GetMeResponse | PlainMessage<GetMeResponse> | undefined, b: GetMeResponse | PlainMessage<GetMeResponse> | undefined): boolean;
}

/**
 * @generated from message bff.v1.JoinApprovalQueueRequest
 */
export declare class JoinApprovalQueueRequest extends Message<JoinApprovalQueueRequest> {
  constructor(data?: PartialMessage<JoinApprovalQueueRequest>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.JoinApprovalQueueRequest";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JoinApprovalQueueRequest;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JoinApprovalQueueRequest;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JoinApprovalQueueRequest;

  static equals(a: JoinApprovalQueueRequest | PlainMessage<JoinApprovalQueueRequest> | undefined, b: JoinApprovalQueueRequest | PlainMessage<JoinApprovalQueueRequest> | undefined): boolean;
}

/**
 * @generated from message bff.v1.JoinApprovalQueueResponse
 */
export declare class JoinApprovalQueueResponse extends Message<JoinApprovalQueueResponse> {
  constructor(data?: PartialMessage<JoinApprovalQueueResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "bff.v1.JoinApprovalQueueResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JoinApprovalQueueResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JoinApprovalQueueResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JoinApprovalQueueResponse;

  static equals(a: JoinApprovalQueueResponse | PlainMessage<JoinApprovalQueueResponse> | undefined, b: JoinApprovalQueueResponse | PlainMessage<JoinApprovalQueueResponse> | undefined): boolean;
}

