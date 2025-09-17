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
  console.log(repos?.data);
  return repos?.data;
};

const validateUser = async () => {
  const resp = await axiosClient.get("/verify/user");
  return resp?.data;
};

const listAllRepos = async () => {
  const resp = await axiosClient.post("/fetch/repos");
  return resp;
};

// const deleteRepos = (){}

export { authenticateUser, validateUser, listAllRepos, searchRepos };
