// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx

import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "./components/auth/AuthProvider";
import { MainLayout } from "./components/layout/MainLayout";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      staleTime: 30_000,
    },
  },
});

/**
 * App — root component.
 *
 * Provider hierarchy:
 *   QueryClientProvider → AuthProvider → MainLayout
 */
const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MainLayout />
      </AuthProvider>
    </QueryClientProvider>
  );
};

export default App;
