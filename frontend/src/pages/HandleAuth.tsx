import useAuth from "@/hooks/useAuth";
import { useAppStore } from "@/store/useAppStore";
import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useShallow } from "zustand/react/shallow";

/** /auth — receives the temporary code & state from GitHub and exchanges them for a session */
const HandleAuth = () => {
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code");
  const state = searchParams.get("state");

  const { setIsAuthenticated, setUser } = useAppStore(
    useShallow((state) => ({
      setIsAuthenticated: state.setIsAuthenticated,
      setUser: state.setUser,
    }))
  );

  const navigate = useNavigate();
  const { data, loading, error } = useAuth(code, state);

  useEffect(() => {
    if (loading) return;

    if (error) {
      console.error("Auth error:", error);
      navigate("/");
      return;
    }

    if (data) {
      const user: User = data.data.user;
      if (!user) {
        console.error("Auth error: no user returned from callback");
        navigate("/");
        return;
      }
      setUser(user);
      setIsAuthenticated(true);
      navigate("/search");
    }
  }, [data, loading, error, navigate, setIsAuthenticated, setUser]);

  if (loading) {
    return <p>Authenticating…</p>;
  }
};

export default HandleAuth;
