import { Transport, createPromiseClient } from "@bufbuild/connect";
import { createConnectTransport } from "@bufbuild/connect-web";
import { ModerationService } from "../../proto/bff/v1/moderation_service_connectweb";
import { UserService } from "../../proto/bff/v1/user_service_connectweb";
import { createRegistry } from "@bufbuild/protobuf";
import {
  AssignRolesAuditPayload,
  BanActorAuditPayload,
  CommentAuditPayload,
  CreateActorAuditPayload,
  ForceApproveActorAuditPayload,
  HoldBackPendingActorAuditPayload,
  ProcessApprovalQueueAuditPayload,
  UnapproveActorAuditPayload,
} from "../../proto/bff/v1/moderation_service_pb";

export async function useAPITransport(): Promise<Transport> {
  const { apiUrl } = useRuntimeConfig().public;
  const user = await useUser();
  return createConnectTransport({
    baseUrl: apiUrl,

    fetch(input, data = {}) {
      (data.headers as Headers).set(
        "authorization",
        `Bearer ${user.value.accessJwt}`
      );

      return globalThis.fetch(input, { ...data });
    },

    jsonOptions: {
      typeRegistry: createRegistry(
        BanActorAuditPayload,
        CommentAuditPayload,
        CreateActorAuditPayload,
        ProcessApprovalQueueAuditPayload,
        UnapproveActorAuditPayload,
        ForceApproveActorAuditPayload,
        HoldBackPendingActorAuditPayload,
        AssignRolesAuditPayload
      ),
    },
  });
}

export default async function () {
  const transport = await useAPITransport();
  return createPromiseClient(ModerationService, transport);
}

export async function useUserAPI() {
  const transport = await useAPITransport();
  return createPromiseClient(UserService, transport);
}
