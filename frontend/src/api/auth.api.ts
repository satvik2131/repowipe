import { AxiosResponse } from "axios";
import axiosClient from "./axiosClient";

type AuthCallbackResponse = {
  user: User;
  connections?: ConnectionsResponse;
  mode?: "login" | "link";
};

const getProviderLoginUrl = async (
  provider: Provider,
  mode: "login" | "link" = "login"
): Promise<string> => {
  const resp = await axiosClient.get<{ url: string }>(
    `/auth/${provider}/login`,
    { params: { mode } }
  );
  return resp.data.url;
};

/** @deprecated use getProviderLoginUrl("github") */
const getGithubLoginUrl = async (): Promise<string> =>
  getProviderLoginUrl("github", "login");

const authenticateUser = async (
  provider: Provider,
  code: string,
  state: string,
  mode: "login" | "link" = "login"
): Promise<AxiosResponse<AuthCallbackResponse>> => {
  return axiosClient.post<AuthCallbackResponse>(
    `/auth/${provider}/callback`,
    { code, state, mode }
  );
};

const validateUser = async (): Promise<boolean> => {
  const resp = await axiosClient.get<boolean>("/verify/user");
  return resp.data;
};

const logoutUser = async (): Promise<void> => {
  await axiosClient.post("/auth/logout");
};

const getConnections = async (): Promise<ConnectionsResponse> => {
  const resp = await axiosClient.get<ConnectionsResponse>("/auth/connections");
  return resp.data;
};

const unlinkProvider = async (
  provider: Provider
): Promise<ConnectionsResponse> => {
  const resp = await axiosClient.delete<ConnectionsResponse>(
    `/auth/${provider}`
  );
  return resp.data;
};

export {
  getProviderLoginUrl,
  getGithubLoginUrl,
  authenticateUser,
  validateUser,
  logoutUser,
  getConnections,
  unlinkProvider,
};
