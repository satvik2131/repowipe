import { AxiosResponse } from "axios";
import axiosClient from "./axiosClient";

const listAllRepos = async (
  provider: Provider,
  page: number,
  filters?: Pick<RepoListFilters, "visibility" | "sort" | "direction">
): Promise<AxiosResponse<Repos[]>> => {
  const params = new URLSearchParams({
    page: String(page),
    visibility: filters?.visibility ?? "all",
    sort: filters?.sort === "stars" ? "updated" : (filters?.sort ?? "updated"),
    direction: filters?.direction ?? "desc",
  });
  return axiosClient.post(`/${provider}/repos?${params.toString()}`);
};

const searchRepos = async (
  provider: Provider,
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

  const resp = await axiosClient.get(`/${provider}/search?${params.toString()}`);
  return resp?.data ?? [];
};

const deleteRepos = async (
  provider: Provider,
  deleteRepoData: DeleteRepoData
): Promise<AxiosResponse> => {
  return axiosClient.delete(`/${provider}/repos`, {
    data: deleteRepoData,
  });
};

export { listAllRepos, searchRepos, deleteRepos };
