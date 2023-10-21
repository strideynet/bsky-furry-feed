import { Actor } from "../../proto/bff/v1/types_pb";
import { useUserAPI } from "./useAPI";

type ActorWithInfo = Actor & {
  isAdmin: boolean;
  isModOrHigher: boolean;
};

export default async function (): Promise<Ref<ActorWithInfo>> {
  const actor = useState<ActorWithInfo>("actor");

  if (!actor.value) {
    const api = await useUserAPI();
    const result = (await api
      .getMe({})
      .then((r) => ({ ...r.Actor }))) as ActorWithInfo;

    result.isAdmin = result?.roles?.includes("admin");
    result.isModOrHigher =
      result.isAdmin || result?.roles?.includes("moderator");

    actor.value = result;
  }

  return actor;
}
