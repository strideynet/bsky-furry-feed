import { ProfileViewDetailed } from "@atproto/api/dist/client/types/app/bsky/actor/defs";
import { newAgent } from "~/lib/auth";

export default async function () {
  const user = await useUser();
  const profile = useState<ProfileViewDetailed>("profile");

  const agent = newAgent();

  if (user.value && !profile.value) {
    const { data } = await agent.getProfile({ actor: user.value.did });
    profile.value = data;
  }

  return profile;
}
