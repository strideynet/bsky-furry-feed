// @generated by protoc-gen-es v1.2.0 with parameter "target=js+dts"
// @generated from file bff/v1/public_service.proto (package bff.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from message bff.v1.ListFeedsRequest
 */
export const ListFeedsRequest = proto3.makeMessageType(
  "bff.v1.ListFeedsRequest",
  [],
);

/**
 * @generated from message bff.v1.ListFeedsResponse
 */
export const ListFeedsResponse = proto3.makeMessageType(
  "bff.v1.ListFeedsResponse",
  () => [
    { no: 1, name: "feeds", kind: "message", T: Feed, repeated: true },
  ],
);

/**
 * @generated from message bff.v1.Feed
 */
export const Feed = proto3.makeMessageType(
  "bff.v1.Feed",
  () => [
    { no: 1, name: "id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "link", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "display_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "priority", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ],
);

