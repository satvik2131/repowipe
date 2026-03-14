import { useEffect, useState } from "react";
import { authenticateUser } from "@/api/auth.api";
import { AxiosResponse } from "axios";

type AuthCallbackData = AxiosResponse<{ user: User }>;

/**
 * useAuth — exchanges the GitHub OAuth temporary code for a session.
 *
 * Calls the backend `/api/auth/github/callback` endpoint which handles the
 * client_secret exchange entirely server-side.
 */
const useAuth = (code: string | null, state: string | null) => {
  const [data, setData] = useState<AuthCallbackData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<unknown>(null);

  useEffect(() => {
    if (!code) return;

    setLoading(true);
    authenticateUser(code, state ?? "")
      .then((res) => setData(res))
      .catch((err) => setError(err))
      .finally(() => setLoading(false));
  }, [code, state]);

  return { data, loading, error };
};

export default useAuth;
