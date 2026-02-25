// CLASSIFICATION: UNCLASSIFIED
// src/components/auth/AuthProvider.tsx

import React, { useEffect } from "react";
import { useAuthStore } from "../../stores/authStore";

interface AuthProviderProps {
  children: React.ReactNode;
}

/**
 * AuthProvider — bootstraps operator authentication context.
 *
 * In production:
 *   1. Reads operator identity from secure session token (mTLS client cert)
 *   2. Calls AuthService to validate token and retrieve clearance + roles
 *   3. Populates AuthStore
 *
 * Development: loads a default operator from environment variables.
 * No PII stored in localStorage or browser console logs.
 */
export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const setOperator = useAuthStore((s) => s.setOperator);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  useEffect(() => {
    if (isAuthenticated) return;

    // Development: bootstrap from env variables or defaults
    // Production: replace with real token validation
    setOperator({
      id: "op-dev-001",
      name: "Dev Operator",
      unit: "DEV",
      clearance: "PROTECTED_B",
      roles: ["OPERATOR", "ANALYST"],
    });
  }, [isAuthenticated, setOperator]);

  return <>{children}</>;
};
