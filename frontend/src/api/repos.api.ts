import { AxiosResponse } from "axios";
import axiosClient from "./axiosClient";

// ── API calls ─────────────────────────────────────────────────────────────────

/**
 * Fetches a paginated list of the authenticated user's repositories.
 */
const listAllRepos = async (
  page: number,
  filters?: Pick<RepoListFilters, "visibility" | "sort" | "direction">
): Promise<AxiosResponse<Repos[]>> => {
  const params = new URLSearchParams({
    page: String(page),
    visibility: filters?.visibility ?? "all",
    sort: filters?.sort === "stars" ? "updated" : (filters?.sort ?? "updated"),
    direction: filters?.direction ?? "desc",
  });
  const resp = await axiosClient.post(`/fetch/repos?${params.toString()}`);
  return resp;
};

/**
 * Searches GitHub repositories for the given username with optional filters.
 */
const searchRepos = async (
  username: string,
  reponame: string,
  filters?: Partial<RepoListFilters>
): Promise<Repos[]> => {
  const params = new URLSearchParams({
    username,
    reponame,
  });
  if (filters?.language) params.set("language", filters.language);
  if (filters?.visibility && filters.visibility !== "all") {
    params.set("visibility", filters.visibility);
  }
  if (filters?.kind && filters.kind !== "all") {
    params.set("kind", filters.kind);
  }
  if (filters?.sort) {
    const searchSort =
      filters.sort === "stars" || filters.sort === "updated"
        ? filters.sort
        : "updated";
    params.set("sort", searchSort);
  }

  const resp = await axiosClient.get(`/search/repo?${params.toString()}`);
  return resp?.data ?? [];
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
