import { createContext } from "react";
import { AtpAgent } from "@atproto/api";

const AuthContext: React.Context<AuthService | undefined> =
  createContext(undefined);

export const AuthProvider = ({ children }) => {
  const authService: AuthService = {
    loggedIn: false,
  };
  authService.atpClient?.session;
  return (
    <AuthContext.Provider value={authService}>{children}</AuthContext.Provider>
  );
};
