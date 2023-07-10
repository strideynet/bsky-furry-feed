// @generated by protoc-gen-connect-web v0.9.0 with parameter "target=js+dts"
// @generated from file proto/bff/v1/moderation_service.proto (package bff.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { GetActorRequest, GetActorResponse, GetApprovalQueueRequest, GetApprovalQueueResponse, ListActorsRequest, ListActorsResponse, PingRequest, PingResponse, ProcessApprovalQueueRequest, ProcessApprovalQueueResponse } from "./moderation_service_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service bff.v1.ModerationService
 */
export declare const ModerationService: {
  readonly typeName: "bff.v1.ModerationService",
  readonly methods: {
    /**
     * @generated from rpc bff.v1.ModerationService.Ping
     */
    readonly ping: {
      readonly name: "Ping",
      readonly I: typeof PingRequest,
      readonly O: typeof PingResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc bff.v1.ModerationService.GetApprovalQueue
     */
    readonly getApprovalQueue: {
      readonly name: "GetApprovalQueue",
      readonly I: typeof GetApprovalQueueRequest,
      readonly O: typeof GetApprovalQueueResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc bff.v1.ModerationService.ProcessApprovalQueue
     */
    readonly processApprovalQueue: {
      readonly name: "ProcessApprovalQueue",
      readonly I: typeof ProcessApprovalQueueRequest,
      readonly O: typeof ProcessApprovalQueueResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc bff.v1.ModerationService.ListActors
     */
    readonly listActors: {
      readonly name: "ListActors",
      readonly I: typeof ListActorsRequest,
      readonly O: typeof ListActorsResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc bff.v1.ModerationService.GetActor
     */
    readonly getActor: {
      readonly name: "GetActor",
      readonly I: typeof GetActorRequest,
      readonly O: typeof GetActorResponse,
      readonly kind: MethodKind.Unary,
    },
  }
};

