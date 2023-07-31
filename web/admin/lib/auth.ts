import * as atproto from "@atproto/api";

export const COOKIE_NAME = "furrylist-bsky-session";
const BSKY_API = "https://bsky.social";

export function newAgent(): atproto.BskyAgent {
  const agent = new atproto.BskyAgent({ service: BSKY_API });

  const user = useState<atproto.AtpSessionData>("user").value;

  if (user) {
    agent.session = user;
  }

  return agent;
}

export async function logout() {
  const agent = newAgent();
  agent.session = undefined;
  useCookie<null>(COOKIE_NAME, { expires: new Date() }).value = null;
  useState("user").value = null;
  window.location.reload();
}

export async function login(
  identifier: string,
  password: string
): Promise<{ error: any; success: boolean }> {
  const agent = newAgent();
  let success: boolean;
  let data: atproto.AtpSessionData;

  try {
    const result = await agent.login({ identifier, password });
    success = result.success;
    data = result.data;
  } catch (error) {
    return { error, success: false };
  }

  if (!success) {
    return { success, error: "Invalid identifier or password" };
  }

  useCookie<atproto.AtpSessionData>(COOKIE_NAME, {
    expires: new Date(Date.now() + 1000 * 60 * 60 * 24 * 30),
  }).value = data;
  useState("user").value = data;

  return { success, error: null };
}

export async function fetchUser(): Promise<atproto.AtpSessionData | null> {
  const cookie = useCookie<atproto.AtpSessionData | null>(COOKIE_NAME, {
    expires: new Date(Date.now() + 1000 * 60 * 60 * 24 * 30),
  });

  if (!cookie.value) {
    return null;
  }

  const agent = newAgent();
  agent.setPersistSessionHandler((evt, session) => {
    switch (evt) {
      case "create":
      case "update":
        if (session) {
          cookie.value = session;
          agent.session = session;
        }
        break;
      case "expired":
        cookie.value = null;
        useState("user").value = null;
        break;
      default:
        break;
    }
  });
  try {
    const { data, success } = await agent.resumeSession(cookie.value);
    if (!success) {
      return null;
    }
    cookie.value = { ...cookie.value, ...data };
  } catch (error) {
    if (!String(error).includes("XRPC ERROR 400: ExpiredToken")) {
      throw error;
    }
  }

  return cookie.value;
}
