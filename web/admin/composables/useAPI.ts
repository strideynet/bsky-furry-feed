import { createPromiseClient } from "@bufbuild/connect";
import { createConnectTransport } from "@bufbuild/connect-web";
import { ModerationService } from "../../proto/bff/v1/moderation_service_connectweb";

export default async function () {
  const { apiUrl } = useRuntimeConfig().public;
  const user = await useUser();
  const transport = createConnectTransport({
    baseUrl: apiUrl,

    fetch(input, data = {}) {
      (data.headers as Headers).set(
        "authorization",
        `Bearer ${user.value.accessJwt}`
      );

      return globalThis.fetch(input, { ...data });
    },
  });

  return createPromiseClient(ModerationService, transport);
}
