import { AxiosResponse } from "axios";
import axiosClient from "./axiosClient";

// ── Types ─────────────────────────────────────────────────────────────────────

/** Payload sent to the backend token-exchange endpoint. */
type TokenExchangePayload = {
  code: string;
  state: string;
};

/** Response from the backend token-exchange endpoint. */
type AuthCallbackResponse = {
  user: User;
};

// ── API calls ─────────────────────────────────────────────────────────────────

/**
 * Fetches the GitHub OAuth authorization URL from the backend.
 * The backend includes client_id, redirect_uri, scope, and a CSRF state token —
 * none of these are handled by the frontend anymore.
 */
const getGithubLoginUrl = async (): Promise<string> => {
  const resp = await axiosClient.get<{ url: string }>("/auth/github/login");
  return resp.data.url;
};

/**
 * Exchanges the temporary GitHub OAuth code for a session.
 * The backend handles the client_secret and access-token exchange.
 */
const authenticateUser = async (
  code: string,
  state: string
): Promise<AxiosResponse<AuthCallbackResponse>> => {
  const payload: TokenExchangePayload = { code, state };
  const resp = await axiosClient.post<AuthCallbackResponse>(
    "/auth/github/callback",
    JSON.stringify(payload)
  );
  return resp;
};

/**
 * Validates whether the current session cookie is still active.
 */
const validateUser = async (): Promise<boolean> => {
  const resp = await axiosClient.get<boolean>("/verify/user");
  return resp.data;
};

/**
 * Logs out: backend clears Redis session, revokes the GitHub token, and
 * expires the session cookie.
 */
const logoutUser = async (): Promise<void> => {
  await axiosClient.post("/auth/logout");
};

export { getGithubLoginUrl, authenticateUser, validateUser, logoutUser };
