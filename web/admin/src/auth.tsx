import React, { createContext, useMemo, useState } from "react";
import { AtpAgent, AtpSessionData, AtpSessionEvent } from "@atproto/api";

export const AgentContext = createContext<{
  agent: AtpAgent;
  session?: AtpSessionData;
}>(undefined!);

export const AgentProvider = (props: { children: React.ReactNode }) => {
  const [session, setSession] = useState<AtpSessionData | undefined>(undefined);
  const agent = useMemo(() => {
    return new AtpAgent({
      service: "https://bsky.social",
      persistSession(_: AtpSessionEvent, s?: AtpSessionData) {
        setSession(s);
      },
    });
  }, []);

  return (
    <AgentContext.Provider
      value={{
        agent: agent,
        session,
      }}
    >
      {props.children}
    </AgentContext.Provider>
  );
};
