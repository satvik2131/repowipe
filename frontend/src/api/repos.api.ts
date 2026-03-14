import { AxiosResponse } from "axios";
import axiosClient from "./axiosClient";

// ── API calls ─────────────────────────────────────────────────────────────────

/**
 * Fetches a paginated list of the authenticated user's repositories.
 */
const listAllRepos = async (page: number): Promise<AxiosResponse> => {
  const resp = await axiosClient.post(`/fetch/repos?page=${page}`);
  return resp;
};

/**
 * Searches GitHub repositories for the given username and repo name query.
 */
const searchRepos = async (
  username: string,
  reponame: string
): Promise<Repos[]> => {
  const resp = await axiosClient.get(
    `/search/repo?username=${username}&reponame=${reponame}`
  );
  return resp?.data;
};

/**
 * Deletes the specified repositories for the authenticated user.
 */
const deleteRepos = async (
  deleteRepoData: DeleteRepoData
): Promise<AxiosResponse> => {
  const resp = await axiosClient.delete("/delete/repos", {
    data: deleteRepoData,
  });
  return resp;
};

export { listAllRepos, searchRepos, deleteRepos };
