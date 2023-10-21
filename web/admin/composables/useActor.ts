import { Actor } from "../../proto/bff/v1/types_pb";

type ActorWithInfo = Actor & {
  isAdmin: boolean;
  isModOrHigher: boolean;
};

export default async function (): Promise<Ref<ActorWithInfo>> {
  const actor = useState<ActorWithInfo>("actor");

  if (!actor.value) {
    const api = await useAPI();
    const user = await useUser();
    const result = (await api
      .getActor({
        did: user.value.did,
      })
      .then((r) => ({ ...r.actor }))) as ActorWithInfo;

    result.isAdmin = result?.roles?.includes("admin");
    result.isModOrHigher =
      result.isAdmin || result?.roles?.includes("moderator");

    actor.value = result;
  }

  return actor;
}
