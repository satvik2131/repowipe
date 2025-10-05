import { AxiosResponse } from "axios";
import axiosClient from "./axiosClient";

type TempCred = {
  code: string;
  state: string;
};

type RespAccessToken = {
  message: string;
};

const authenticateUser = async (
  code: string,
  state: string
): Promise<RespAccessToken> => {
  const payload: TempCred = {
    code: code,
    state: state,
  };

  try {
    const resp: RespAccessToken = await axiosClient.post(
      "/set/access/token",
      JSON.stringify(payload)
    );
    return resp;
  } catch (error) {
    return error;
  }
};

const searchRepos = async (username: string, reponame: string) => {
  const repos = await axiosClient.get(
    `/search/repo?username=${username}&reponame=${reponame}`
  );
  return repos?.data;
};

const validateUser = async () => {
  const resp = await axiosClient.get("/verify/user");
  return resp?.data;
};

const listAllRepos = async (page: number):Promise<AxiosResponse> => {
  const resp = await axiosClient.post(`/fetch/repos?page=${page}`);
  return resp;
};

const deleteRepos = async (deleteRepos: DeleteRepoData):Promise<AxiosResponse> => {
  const resp = await axiosClient.delete("/delete/repos", {
    data: deleteRepos,
  });
  return resp 
};

export {
  authenticateUser,
  validateUser,
  listAllRepos,
  searchRepos,
  deleteRepos,
};
