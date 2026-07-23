import { Navigate } from "react-router-dom";
import { useAppStore } from "@/store/useAppStore";
import { ReactNode } from "react";

type ProtectedRouteProps = {
  children: ReactNode;
};

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const isAuthenticated = useAppStore((state) => state.isAuthenticated);
  const user = useAppStore((state) => state.user);

  return isAuthenticated && user ? (
    <div className="protected-content">{children}</div>
  ) : (
    <Navigate to="/" replace />
  );
};
