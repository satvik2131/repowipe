import axiosClient from "./axiosClient";

const startTransfer = async (payload: {
  source: Provider;
  destination: Provider;
  repos: string[];
}): Promise<TransferJob> => {
  const resp = await axiosClient.post<TransferJob>("/transfers", payload);
  return resp.data;
};

const getTransfer = async (id: string): Promise<TransferJob> => {
  const resp = await axiosClient.get<TransferJob>(`/transfers/${id}`);
  return resp.data;
};

export { startTransfer, getTransfer };
