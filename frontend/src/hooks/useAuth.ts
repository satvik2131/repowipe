import { useEffect, useState } from "react";
import { authenticateUser } from "@/api/auth.api";
import { AxiosResponse } from "axios";

type AuthCallbackData = AxiosResponse<{
  user: User;
  connections?: ConnectionsResponse;
  mode?: "login" | "link";
}>;

const useAuth = (
  provider: Provider | null,
  code: string | null,
  state: string | null,
  mode: "login" | "link" = "login"
) => {
  const [data, setData] = useState<AuthCallbackData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<unknown>(null);

  useEffect(() => {
    if (!code || !provider) return;

    setLoading(true);
    authenticateUser(provider, code, state ?? "", mode)
      .then((res) => setData(res))
      .catch((err) => setError(err))
      .finally(() => setLoading(false));
  }, [provider, code, state, mode]);

  return { data, loading, error };
};

export default useAuth;
