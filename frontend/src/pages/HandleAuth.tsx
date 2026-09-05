import useAuth from "@/hooks/useAuth";
import { useAppStore } from "@/store/useAppStore";
import { useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useShallow } from "zustand/react/shallow";

const isProvider = (v: string | null | undefined): v is Provider =>
  v === "github" || v === "gitlab" || v === "bitbucket";

/** Parse provider + mode from state: `{provider}:{mode}:{uuid}` */
const parseState = (state: string | null) => {
  if (!state) return { provider: null as Provider | null, mode: "login" as const };
  const parts = state.split(":");
  if (parts.length >= 3 && isProvider(parts[0])) {
    const mode = parts[1] === "link" ? ("link" as const) : ("login" as const);
    return { provider: parts[0], mode };
  }
  return { provider: null as Provider | null, mode: "login" as const };
};

/** /auth — OAuth callback; provider/mode come from state (or query fallback) */
const HandleAuth = () => {
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code");
  const state = searchParams.get("state");
  const fromState = useMemo(() => parseState(state), [state]);
  const providerParam = searchParams.get("provider");
  const modeParam = searchParams.get("mode");

  const provider = isProvider(providerParam)
    ? providerParam
    : fromState.provider ?? "github";
  const mode =
    modeParam === "link" || fromState.mode === "link" ? "link" : "login";

  const { setIsAuthenticated, setUser, setConnections, setActiveProvider } =
    useAppStore(
      useShallow((s) => ({
        setIsAuthenticated: s.setIsAuthenticated,
        setUser: s.setUser,
        setConnections: s.setConnections,
        setActiveProvider: s.setActiveProvider,
      }))
    );

  const navigate = useNavigate();
  const { data, loading, error } = useAuth(provider, code, state, mode);

  useEffect(() => {
    if (loading) return;

    if (error) {
      console.error("Auth error:", error);
      navigate(mode === "link" ? "/search" : "/");
      return;
    }

    if (data) {
      const user: User = data.data.user;
      if (!user && mode === "login") {
        console.error("Auth error: no user returned from callback");
        navigate("/");
        return;
      }
      if (user) {
        setUser(user);
        setIsAuthenticated(true);
        if (mode === "login") {
          setActiveProvider(user.provider ?? provider);
        }
      }
      if (data.data.connections) {
        setConnections(data.data.connections);
        if (mode === "login" && data.data.connections.primary) {
          setActiveProvider(data.data.connections.primary);
        }
      }
      navigate("/search");
    }
  }, [
    data,
    loading,
    error,
    navigate,
    setIsAuthenticated,
    setUser,
    setConnections,
    setActiveProvider,
    mode,
    provider,
  ]);

  if (loading) {
    return <p>Authenticating…</p>;
  }
  return null;
};

export default HandleAuth;
