import { useEffect, useState } from "react";
import { authenticateUser } from "@/api/apis_routes";

const useCallApi = (code: string, state: string) => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>(null);

  useEffect(() => {
    if (!code) {
      return;
    }
    setLoading(true);
    authenticateUser(code, state)
      .then((res) => setData(res))
      .catch((err) => setError(err))
      .finally(() => setLoading(false));
  }, [code, state]);

  return { data, loading, error };
};

export default useCallApi;
