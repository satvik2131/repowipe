import useCallApi from "@/hooks/useCallApi";
import { useAppStore } from "@/store/useAppStore";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useSearchParams } from "react-router-dom";
import { useShallow } from "zustand/react/shallow";
/** /auth - handles temporary token */
const HandleAuth = () => {
  const [searchParams, _] = useSearchParams();
  const token = searchParams.get("code");
  const state = searchParams.get("state");
  const { setIsAuthenticated, setUser } = useAppStore(
    useShallow((state) => ({
      setIsAuthenticated: state.setIsAuthenticated,
      setUser: state.setUser,
    }))
  );
  const navigate = useNavigate();
  const { data, loading, error } = useCallApi(token, state);

  useEffect(() => {
    if (loading) return; // don't navigate while still loading

    if (error) {
      console.error("API error:", error);
      navigate("/"); // failed → go home
      return;
    }

    if (data) {
      const user: User = data.data.user;
      if (user) {
        setUser(user);
      }
      // success case
      setIsAuthenticated(true);
      navigate("/search");
    }
  }, [data, loading, error, navigate, setIsAuthenticated]);

  if (loading) {
    return <p>Authenticating...</p>;
  }
};

export default HandleAuth;
