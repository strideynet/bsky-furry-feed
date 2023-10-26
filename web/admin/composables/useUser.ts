import { AtpSessionData } from "@atproto/api";
import { COOKIE_NAME, fetchUser } from "~/lib/auth";

export default async function (): Promise<Ref<AtpSessionData>> {
  const user = useState("user");
  const session = useCookie(COOKIE_NAME).value;

  if (session && !user.value) {
    user.value = await fetchUser();
  }

  return user as Ref<AtpSessionData>;
}

/**
 * 6.23
12 XHR requests

3.76
7 XHR requests
 */
