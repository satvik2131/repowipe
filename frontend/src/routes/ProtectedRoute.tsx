import { Navigate } from "react-router-dom";
import { useAppStore } from "@/store/useAppStore";
import { ReactNode } from "react";

type ProtectedRouteProps = {
  children: ReactNode;
};

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const isAuthenticated = useAppStore((state) => state.isAuthenticated);

  return isAuthenticated ? (
    <div className="protected-content">{children}</div>
  ) : (
    <Navigate to="/" replace />
  );
};
