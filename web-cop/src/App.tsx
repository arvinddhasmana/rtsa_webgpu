// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx

import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "./components/auth/AuthProvider";
import { MainLayout } from "./components/layout/MainLayout";
import { useUIStore } from "./stores/uiStore";
import "./styles/nvg-theme.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      staleTime: 30_000,
    },
  },
});

/**
 * AppRoot applies the active theme class before rendering the layout.
 */
const AppRoot: React.FC = () => {
  const theme = useUIStore((s) => s.theme);
  return (
    <div className={`app-root theme-${theme}`} style={{ height: "100%" }}>
      <MainLayout />
    </div>
  );
};

/**
 * App — root component.
 *
 * Provider hierarchy:
 *   QueryClientProvider → AuthProvider → AppRoot → MainLayout
 */
const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <AppRoot />
      </AuthProvider>
    </QueryClientProvider>
  );
};

export default App;
